package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ilpaka/landmark_app/backend/trips-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/adapters/postgres"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/app"
	triphttp "github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg := config.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("pg", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := postgres.New(pool)
	svc := app.New(store, cfg.OSRMUrl)
	stack := triphttp.Mount(svc)

	log.Info("starting trips-service", "addr", cfg.HTTPAddr)
	if err := stack.Engine.Run(cfg.HTTPAddr); err != nil {
		log.Error("server fatal", "err", err)
		os.Exit(1)
	}
}
