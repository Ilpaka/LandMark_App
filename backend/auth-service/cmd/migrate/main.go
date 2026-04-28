package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|status>")
		os.Exit(2)
	}
	cmd := os.Args[1]

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("dialect: %v", err)
	}

	dir := os.Getenv("GOOSE_MIGRATION_DIR")
	if dir == "" {
		dir = "db/migrations"
	}

	switch cmd {
	case "up":
		if err := goose.Up(db, dir); err != nil {
			log.Fatalf("goose up: %v", err)
		}
	case "down":
		if err := goose.Down(db, dir); err != nil {
			log.Fatalf("goose down: %v", err)
		}
	case "status":
		if err := goose.Status(db, dir); err != nil {
			log.Fatalf("goose status: %v", err)
		}
	default:
		log.Fatal("unknown command (use up, down, status)")
	}
}
