package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	dir := os.Getenv("GOOSE_MIGRATION_DIR")
	if dir == "" {
		dir = "db/migrations"
	}
	goose.SetBaseFS(nil)
	if err := goose.Up(db, dir); err != nil {
		log.Error("migration", "err", err)
		os.Exit(1)
	}
}
