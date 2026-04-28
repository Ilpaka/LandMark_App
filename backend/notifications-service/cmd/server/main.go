package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/adapters/postgres"
	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/app"
	notifhttp "github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/http"
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
	svc := &app.Service{Store: store}
	h := &notifhttp.Handler{Svc: svc}

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	notifhttp.Mount(r, h)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Info("notifications-service started", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Error("server error", "err", err)
		os.Exit(1)
	}
}
