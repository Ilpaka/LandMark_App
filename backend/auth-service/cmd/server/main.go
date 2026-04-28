package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/httpserver"
	redisx "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/redis"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/worker"
	"github.com/ilpaka/landmark_app/backend/pkg/observability"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	otelShutdown, err := observability.SetupOTel(ctx, cfg.OTELServiceName, cfg.OTELExporterOTLPEndpoint)
	if err != nil {
		log.Error("otel setup", "err", err)
		os.Exit(1)
	}
	defer func() {
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(sctx); err != nil {
			log.Warn("otel shutdown", "err", err)
		}
	}()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("pg connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb, err := redisx.New(cfg.RedisURL)
	if err != nil {
		log.Error("redis", "err", err)
		os.Exit(1)
	}

	stack, err := httpserver.MountAuth(cfg, log, pool, rdb)
	if err != nil {
		log.Error("http mount", "err", err)
		os.Exit(1)
	}

	srvCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.OutboxPoller(srvCtx, stack.Store, log, 2*time.Second)

	go func() {
		if err := stack.Engine.Run(cfg.HTTPAddr); err != nil {
			log.Error("http", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	cancel()
}
