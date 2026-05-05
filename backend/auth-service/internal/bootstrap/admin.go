// Package bootstrap performs one-shot startup tasks for auth-service.
//
// EnsureAdmin reconciles the demo/admin account on each start: if no account
// exists for the given email, it creates one with role=admin, email already
// verified and status=active. If an account exists, the role is promoted to
// admin. The operation is idempotent; calling it on every start has no
// observable effect after the first successful run.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
)

// EnsureAdmin makes sure an account with the given email exists and has the
// admin role. The account is created with status='active' and a verified email
// timestamp set to now. Password is hashed with the supplied Argon2id hasher.
//
// Pass empty email or password to skip (used to disable bootstrap in prod).
func EnsureAdmin(
	ctx context.Context,
	pool *pgxpool.Pool,
	hasher *crypto.Argon2idHasher,
	email, password string,
) error {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return nil
	}
	emailNorm := strings.ToLower(email)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("admin bootstrap: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var (
		id   uuid.UUID
		role string
	)
	err = tx.QueryRow(ctx, `
		SELECT id, role FROM auth_accounts
		WHERE email_normalized = $1 AND deleted_at IS NULL
	`, emailNorm).Scan(&id, &role)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		hash, herr := hasher.Hash(password)
		if herr != nil {
			return fmt.Errorf("admin bootstrap: hash password: %w", herr)
		}
		now := time.Now().UTC()
		err = tx.QueryRow(ctx, `
			INSERT INTO auth_accounts
				(email, email_normalized, status, role, email_verified_at, created_at, updated_at)
			VALUES ($1, $2, 'active', 'admin', $3, $3, $3)
			RETURNING id
		`, email, emailNorm, now).Scan(&id)
		if err != nil {
			return fmt.Errorf("admin bootstrap: insert account: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO auth_credentials
				(user_id, password_hash, algo_version, password_changed_at)
			VALUES ($1, $2, 1, $3)
		`, id, hash, now)
		if err != nil {
			return fmt.Errorf("admin bootstrap: insert credentials: %w", err)
		}

	case err != nil:
		return fmt.Errorf("admin bootstrap: lookup: %w", err)

	default:
		// Account exists; promote to admin if not already, ensure verified.
		if role != "admin" {
			_, err = tx.Exec(ctx, `
				UPDATE auth_accounts
				   SET role='admin', updated_at=now()
				 WHERE id=$1
			`, id)
			if err != nil {
				return fmt.Errorf("admin bootstrap: promote: %w", err)
			}
		}
		// Always make sure status is active and email verified for the demo
		// admin (so a previously blocked admin can be unblocked by restart).
		_, err = tx.Exec(ctx, `
			UPDATE auth_accounts
			   SET status='active',
			       blocked_at=NULL, blocked_reason=NULL,
			       email_verified_at=COALESCE(email_verified_at, now()),
			       updated_at=now()
			 WHERE id=$1
		`, id)
		if err != nil {
			return fmt.Errorf("admin bootstrap: reactivate: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("admin bootstrap: commit: %w", err)
	}
	return nil
}
