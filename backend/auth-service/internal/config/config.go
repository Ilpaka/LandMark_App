package config

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration loaded from environment.
type Config struct {
	Env                string
	HTTPAddr           string
	DatabaseURL        string
	RedisURL           string
	JWTPrivateKeyPath  string
	JWTPublicKeyPath   string
	JWTIssuer          string
	JWTAudience        string
	Pepper             string
	ProfileServiceURL  string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	VerificationOTPTTL time.Duration
	ResetOTPTTL        time.Duration
	ArgonMemoryKiB     uint32
	ArgonIterations    uint32
	ArgonParallelism   uint8
	// DevSwagger, when true with APP_ENV=production, enables GET /docs and OpenAPI routes (otherwise they are on by default outside production).
	DevSwagger bool
	// OpenAPIAuthYAML optional absolute or relative path to auth.yaml; if empty, the server searches default locations (service module root or repo root).
	OpenAPIAuthYAML string
	// OTELServiceName is the OpenTelemetry service.name resource attribute (otelgin span scope).
	OTELServiceName string
	// OTELExporterOTLPEndpoint is the OTLP/HTTP traces endpoint (e.g. http://127.0.0.1:4318 or .../v1/traces). Empty disables export.
	OTELExporterOTLPEndpoint string
	// SMSRUBAPIID is the api_id from https://sms.ru/ ; empty uses in-process log stub in MountAuth.
	SMSRUBAPIID       string
	SMSRUBHTTPTimeout time.Duration
}

func Load() (*Config, error) {
	c := &Config{
		Env:                      getenv("APP_ENV", "development"),
		HTTPAddr:                 getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		RedisURL:                 getenv("REDIS_URL", "redis://127.0.0.1:6379/0"),
		JWTPrivateKeyPath:        os.Getenv("JWT_PRIVATE_KEY_PATH"),
		JWTPublicKeyPath:         os.Getenv("JWT_PUBLIC_KEY_PATH"),
		JWTIssuer:                getenv("JWT_ISSUER", "landmark.app"),
		JWTAudience:              getenv("JWT_AUDIENCE", "landmark-clients"),
		Pepper:                   os.Getenv("AUTH_PEPPER"),
		ProfileServiceURL:        os.Getenv("PROFILE_SERVICE_URL"),
		AccessTokenTTL:           durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:          durationEnv("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		VerificationOTPTTL:       durationEnv("VERIFICATION_OTP_TTL", 15*time.Minute),
		ResetOTPTTL:              durationEnv("RESET_OTP_TTL", 15*time.Minute),
		ArgonMemoryKiB:           clampUint32(uintEnv("AUTH_ARGON_MEMORY_KIB", 65536)), // 64 MiB
		ArgonIterations:          clampUint32(uintEnv("AUTH_ARGON_TIME", 3)),
		ArgonParallelism:         clampUint8(uintEnv("AUTH_ARGON_PARALLELISM", 2)),
		DevSwagger:               strings.EqualFold(os.Getenv("DEV_SWAGGER"), "true"),
		OpenAPIAuthYAML:          os.Getenv("OPENAPI_AUTH_YAML"),
		OTELServiceName:          getenv("OTEL_SERVICE_NAME", "landmark-auth-api"),
		OTELExporterOTLPEndpoint: strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		SMSRUBAPIID:              strings.TrimSpace(os.Getenv("SMSRU_API_ID")),
		SMSRUBHTTPTimeout:        durationEnv("SMSRU_HTTP_TIMEOUT", 15*time.Second),
	}
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.Pepper == "" {
		return nil, fmt.Errorf("AUTH_PEPPER is required")
	}
	if c.JWTPrivateKeyPath == "" || c.JWTPublicKeyPath == "" {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH are required")
	}
	return c, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func durationEnv(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func uintEnv(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func clampUint32(n int) uint32 {
	if n < 0 {
		return 0
	}
	if n > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(n) //nolint:gosec // bounded by checks above
}

func clampUint8(n int) uint8 {
	if n < 0 {
		return 0
	}
	if n > math.MaxUint8 {
		return math.MaxUint8
	}
	return uint8(n) //nolint:gosec // bounded by checks above
}
