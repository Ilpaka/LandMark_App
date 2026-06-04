package httpserver

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/email"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/jwt"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/postgres"
	profileadapter "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/profile"
	redisx "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/redis"
	smsadapter "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/sms"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/app"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	authhttp "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/http"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// AuthStack holds the HTTP engine, auth service, and store (e.g. for background workers).
type AuthStack struct {
	Engine *gin.Engine
	Svc    *app.Service
	Store  *postgres.Store
}

// MountAuth wires Gin with health, metrics, JWKS, and /v1 auth routes.
func MountAuth(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool, rdb *redisx.Client) (*AuthStack, error) {
	signer, err := jwt.LoadSigner(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath, "", cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		return nil, err
	}

	var profile ports.ProfileClient = profileadapter.Stub{}
	if cfg.ProfileServiceURL != "" {
		profile = &profileadapter.HTTPClient{BaseURL: cfg.ProfileServiceURL}
	}

	var mail ports.EmailSender = email.LogSender{Log: log}
	if smtp := email.NewFromEnv(log); smtp.Cfg.Addr != "" {
		mail = smtp
	}

	var sms ports.SMSSender = smsadapter.LogSender{Log: log}
	if cfg.SMSRUBAPIID != "" {
		sms = &smsadapter.SMSruSender{
			APIID:   cfg.SMSRUBAPIID,
			Timeout: cfg.SMSRUBHTTPTimeout,
		}
	}

	hasher := &crypto.Argon2idHasher{
		Pepper:      cfg.Pepper,
		MemoryKiB:   cfg.ArgonMemoryKiB,
		Time:        cfg.ArgonIterations,
		Parallelism: cfg.ArgonParallelism,
		SaltLength:  16,
		KeyLength:   32,
	}

	pgStore := postgres.New(pool)
	svc, err := app.NewService(
		pgStore,
		rdb,
		rdb,
		hasher,
		signer,
		profile,
		mail,
		sms,
		cfg.Pepper,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.VerificationOTPTTL,
		cfg.ResetOTPTTL,
		log,
	)
	if err != nil {
		return nil, err
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(RequestIDMiddleware())
	engine.Use(otelgin.Middleware(cfg.OTELServiceName))
	engine.Use(PrometheusHTTPMiddleware())

	engine.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "db": err.Error()})
			return
		}
		if err := rdb.RDB().Ping(ctx).Err(); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "redis": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	engine.GET("/.well-known/jwks.json", func(c *gin.Context) {
		raw, err := signer.JWKS(c.Request.Context())
		if err != nil {
			c.Status(500)
			return
		}
		c.Data(200, "application/json", raw)
	})

	v1 := engine.Group("/v1")
	authhttp.Mount(v1, svc, signer, rdb)

	if shouldExposeSwagger(cfg) {
		mountSwaggerUI(engine, findAuthOpenAPISpec(cfg))
	}

	return &AuthStack{Engine: engine, Svc: svc, Store: pgStore}, nil
}
