package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/adapters/postgres"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/app"
	journalhttp "github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/http"
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
	svc := app.New(store)
	stack := journalhttp.Mount(svc)

	log.Info("starting journal-service", "addr", cfg.HTTPAddr)
	if err := stack.Engine.Run(cfg.HTTPAddr); err != nil {
		log.Error("server fatal", "err", err)
		os.Exit(1)
	}
}
