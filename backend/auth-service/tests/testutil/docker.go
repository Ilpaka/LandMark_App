//go:build integration || e2e

package testutil

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/pressly/goose/v3"

	redisx "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/redis"
)

// RepoRoot returns the Tennis_app repository root (directory containing backend/, app/, docs/).
func RepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	// backend/auth-service/tests/testutil/docker.go -> repo root is ../../../../
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}

// BackendRoot returns the auth-service module directory (Go module root).
func BackendRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join(RepoRoot(t), "backend", "auth-service")
}

// MigrationsDir returns the goose migrations path.
func MigrationsDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(BackendRoot(t), "db", "migrations")
}

// WriteTestRSAKeys writes a PKCS1 RSA keypair into dir and returns paths.
func WriteTestRSAKeys(t *testing.T, dir string) (privPath, pubPath string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	privPath = filepath.Join(dir, "jwt_rsa.pem")
	pubPath = filepath.Join(dir, "jwt_rsa.pub.pem")
	if err := os.WriteFile(privPath, privBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pubPath, pubBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	return privPath, pubPath
}

// RunGooseUp applies all migrations to dsn.
func RunGooseUp(t *testing.T, dsn, migrationsDir string) {
	t.Helper()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
}

// AuthDockerDeps holds resources created by RequireAuthDocker.
type AuthDockerDeps struct {
	DSN      string
	RedisURL string
	Pool     *pgxpool.Pool
	RDB      *redisx.Client
	PrivPath string
	PubPath  string
}

// RequireAuthDocker starts Postgres 16 + Redis 7, runs migrations, generates JWT keys,
// wires DATABASE_URL and REDIS_URL via t.Setenv, and returns a connected pool and Redis client.
// Skips the test if Docker is unavailable.
func RequireAuthDocker(t *testing.T) *AuthDockerDeps {
	t.Helper()
	dp, err := dockertest.NewPool("")
	if err != nil {
		t.Skipf("docker pool: %v", err)
	}
	if err := dp.Client.Ping(); err != nil {
		t.Skipf("docker not available: %v", err)
	}

	pgRes, err := dp.Run("postgres", "16-alpine", []string{
		"POSTGRES_USER=landmark",
		"POSTGRES_PASSWORD=landmark",
		"POSTGRES_DB=landmark",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pgRes.Close() })
	_ = pgRes.Expire(180)

	dsn := fmt.Sprintf("postgres://landmark:landmark@%s/landmark?sslmode=disable&search_path=auth,public", pgRes.GetHostPort("5432/tcp"))
	if err := dp.Retry(func() error {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	}); err != nil {
		t.Fatal(err)
	}

	redisRes, err := dp.Run("redis", "7-alpine", []string{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = redisRes.Close() })
	_ = redisRes.Expire(180)

	redisURL := "redis://" + redisRes.GetHostPort("6379/tcp") + "/0"
	if err := dp.Retry(func() error {
		r, err := redisx.New(redisURL)
		if err != nil {
			return err
		}
		return r.RDB().Ping(context.Background()).Err()
	}); err != nil {
		t.Fatal(err)
	}

	RunGooseUp(t, dsn, MigrationsDir(t))

	keyDir := t.TempDir()
	privPath, pubPath := WriteTestRSAKeys(t, keyDir)

	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("REDIS_URL", redisURL)
	t.Setenv("AUTH_PEPPER", "integration-test-pepper-not-for-prod-32b")
	t.Setenv("JWT_PRIVATE_KEY_PATH", privPath)
	t.Setenv("JWT_PUBLIC_KEY_PATH", pubPath)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	rdb, err := redisx.New(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rdb.RDB().Close() })

	return &AuthDockerDeps{
		DSN:      dsn,
		RedisURL: redisURL,
		Pool:     pool,
		RDB:      rdb,
		PrivPath: privPath,
		PubPath:  pubPath,
	}
}
