package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	placesclient "github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/adapters/places"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/adapters/postgres"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/app"
	modhttp "github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/http"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Error("DATABASE_URL required")
		os.Exit(1)
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := &postgres.Store{Pool: pool}
	placesClient := placesclient.New()
	svc := &app.Service{Store: store, Places: placesClient}
	h := &modhttp.Handler{Svc: svc}

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	modhttp.Mount(r, h)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Info("moderation-service started", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Error("server error", "err", err)
		os.Exit(1)
	}
}
