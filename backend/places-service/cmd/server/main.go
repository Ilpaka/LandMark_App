package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/adapters/moderation"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/adapters/postgres"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/app"
	placehttp "github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/http"
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
	modClient := moderation.NewClient(cfg.ModerationServiceURL, cfg.InternalAPIKey)
	svc := app.NewWithModeration(store, modClient)
	stack := placehttp.Mount(svc, cfg.InternalAPIKey)

	log.Info("starting places-service", "addr", cfg.HTTPAddr)
	if err := stack.Engine.Run(cfg.HTTPAddr); err != nil {
		log.Error("server fatal", "err", err)
		os.Exit(1)
	}
}
