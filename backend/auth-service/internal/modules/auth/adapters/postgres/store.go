package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// Store implements ports.Store using PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Pool exposes the connection pool for health checks.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) WithTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) GetAccountByEmailNorm(ctx context.Context, tx pgx.Tx, emailNorm string) (*ports.Account, error) {
	row := tx.QueryRow(ctx, `
SELECT id, email, email_normalized, phone_e164, phone_verified_at, status, role, email_verified_at, last_login_at,
       blocked_at, blocked_reason, deleted_at, created_at, updated_at
FROM auth_accounts WHERE email_normalized=$1 AND deleted_at IS NULL`, emailNorm)
	return scanAccount(row)
}

func (s *Store) GetAccountByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*ports.Account, error) {
	row := tx.QueryRow(ctx, `
SELECT id, email, email_normalized, phone_e164, phone_verified_at, status, role, email_verified_at, last_login_at,
       blocked_at, blocked_reason, deleted_at, created_at, updated_at
FROM auth_accounts WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanAccount(row)
}

func (s *Store) GetAccountByPhoneE164(ctx context.Context, tx pgx.Tx, phoneE164 string) (*ports.Account, error) {
	row := tx.QueryRow(ctx, `
SELECT id, email, email_normalized, phone_e164, phone_verified_at, status, role, email_verified_at, last_login_at,
       blocked_at, blocked_reason, deleted_at, created_at, updated_at
FROM auth_accounts WHERE phone_e164=$1 AND deleted_at IS NULL`, phoneE164)
	return scanAccount(row)
}

// ListAccounts returns active and blocked accounts (deleted_at IS NULL),
// optionally filtered by a fragment matched against email_normalized or
// phone_e164. Limit is clamped to [1,200], offset to [0, +inf).
func (s *Store) ListAccounts(
	ctx context.Context, tx pgx.Tx,
	query string, limit, offset int,
) ([]ports.Account, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	q := strings.TrimSpace(strings.ToLower(query))

	rows, err := tx.Query(ctx, `
SELECT id, email, email_normalized, phone_e164, phone_verified_at, status, role, email_verified_at, last_login_at,
       blocked_at, blocked_reason, deleted_at, created_at, updated_at
FROM auth_accounts
WHERE deleted_at IS NULL
  AND ($1 = '' OR email_normalized LIKE '%' || $1 || '%' OR phone_e164 LIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ports.Account, 0, limit)
	for rows.Next() {
		acc, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *acc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func nullStrPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

func nullTimePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time
	return &t
}

func scanAccount(row pgx.Row) (*ports.Account, error) {
	var a ports.Account
	var status, role string
	var email, emailNorm, phoneE164 sql.NullString
	var phoneVer sql.NullTime
	if err := row.Scan(
		&a.ID, &email, &emailNorm, &phoneE164, &phoneVer, &status, &role,
		&a.EmailVerifiedAt, &a.LastLoginAt, &a.BlockedAt, &a.BlockedReason, &a.DeletedAt,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	a.Email = nullStrPtr(email)
	a.EmailNormalized = nullStrPtr(emailNorm)
	a.PhoneE164 = nullStrPtr(phoneE164)
	a.PhoneVerifiedAt = nullTimePtr(phoneVer)
	a.Status = domain.AccountStatus(status)
	a.Role = domain.Role(role)
	return &a, nil
}

func (s *Store) InsertAccount(ctx context.Context, tx pgx.Tx, email, emailNorm string, status domain.AccountStatus, role domain.Role) (*ports.Account, error) {
	var e, n any
	if email != "" {
		e = email
	}
	if emailNorm != "" {
		n = emailNorm
	}
	row := tx.QueryRow(ctx, `
INSERT INTO auth_accounts (email, email_normalized, status, role)
VALUES ($1,$2,$3,$4)
RETURNING id, email, email_normalized, phone_e164, phone_verified_at, status, role, email_verified_at, last_login_at,
          blocked_at, blocked_reason, deleted_at, created_at, updated_at`,
		e, n, string(status), string(role))
	return scanAccount(row)
}

func (s *Store) InsertAccountPhone(ctx context.Context, tx pgx.Tx, phoneE164 string, status domain.AccountStatus, role domain.Role) (*ports.Account, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO auth_accounts (phone_e164, status, role) VALUES ($1,$2,$3)
RETURNING id, email, email_normalized, phone_e164, phone_verified_at, status, role, email_verified_at, last_login_at,
          blocked_at, blocked_reason, deleted_at, created_at, updated_at`,
		phoneE164, string(status), string(role))
	return scanAccount(row)
}

func (s *Store) MarkAccountPhoneVerified(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_accounts
SET phone_verified_at=$2, status='active', updated_at=now()
WHERE id=$1 AND deleted_at IS NULL`, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) SetAccountPendingEmail(ctx context.Context, tx pgx.Tx, id uuid.UUID, email, emailNorm string) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_accounts SET email=$2, email_normalized=$3, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`,
		id, email, emailNorm)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) UpdateAccountStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status domain.AccountStatus) error {
	tag, err := tx.Exec(ctx, `UPDATE auth_accounts SET status=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id, string(status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) SetEmailVerified(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_accounts
SET email_verified_at=$2, status=CASE WHEN status='pending_verification' THEN 'active' ELSE status END, updated_at=now()
WHERE id=$1 AND deleted_at IS NULL`,
		id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) SetBlocked(ctx context.Context, tx pgx.Tx, id uuid.UUID, reason string, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_accounts SET status='blocked', blocked_at=$2, blocked_reason=$3, updated_at=now()
WHERE id=$1 AND deleted_at IS NULL`, id, at, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) ClearBlocked(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_accounts SET status='active', blocked_at=NULL, blocked_reason=NULL, updated_at=now()
WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) SoftDeleteAccount(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_accounts SET status='deleted', deleted_at=$2, updated_at=now() WHERE id=$1`, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) UpdateLastLogin(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error {
	_, err := tx.Exec(ctx, `UPDATE auth_accounts SET last_login_at=$2, updated_at=now() WHERE id=$1`, id, at)
	return err
}

func (s *Store) UpsertCredential(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string, algo int16) error {
	_, err := tx.Exec(ctx, `
INSERT INTO auth_credentials (user_id, password_hash, algo_version, password_changed_at, failed_login_count, last_failed_login_at)
VALUES ($1,$2,$3,now(),0,NULL)
ON CONFLICT (user_id) DO UPDATE SET password_hash=EXCLUDED.password_hash, algo_version=EXCLUDED.algo_version, password_changed_at=now()`,
		userID, passwordHash, algo)
	return err
}

func (s *Store) GetCredential(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*ports.Credential, error) {
	row := tx.QueryRow(ctx, `
SELECT user_id, password_hash, algo_version, password_changed_at, failed_login_count, last_failed_login_at
FROM auth_credentials WHERE user_id=$1`, userID)
	var c ports.Credential
	if err := row.Scan(&c.UserID, &c.PasswordHash, &c.AlgoVersion, &c.PasswordChangedAt, &c.FailedLoginCount, &c.LastFailedLoginAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (s *Store) UpdateCredentialFailures(ctx context.Context, tx pgx.Tx, userID uuid.UUID, count int, lastFail *time.Time) error {
	_, err := tx.Exec(ctx, `
UPDATE auth_credentials SET failed_login_count=$2, last_failed_login_at=$3 WHERE user_id=$1`,
		userID, count, lastFail)
	return err
}

func (s *Store) ResetCredentialFailures(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
UPDATE auth_credentials SET failed_login_count=0, last_failed_login_at=NULL WHERE user_id=$1`, userID)
	return err
}

func (s *Store) UpdatePasswordHash(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string, algo int16, changedAt time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_credentials SET password_hash=$2, algo_version=$3, password_changed_at=$4 WHERE user_id=$1`,
		userID, passwordHash, algo, changedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) CancelPendingVerifications(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
UPDATE auth_email_verifications SET status='cancelled' WHERE user_id=$1 AND status='pending'`, userID)
	return err
}

func (s *Store) InsertVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, email string, salt []byte, codeHash string, expires time.Time) (*ports.EmailVerification, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO auth_email_verifications (user_id, email, code_salt, code_hash, expires_at, attempts_count, max_attempts, status)
VALUES ($1,$2,$3,$4,$5,0,5,'pending')
RETURNING id, user_id, email, code_salt, code_hash, expires_at, attempts_count, max_attempts, status, verified_at, created_at`,
		userID, email, salt, codeHash, expires)
	return scanVerification(row)
}

func scanVerification(row pgx.Row) (*ports.EmailVerification, error) {
	var v ports.EmailVerification
	if err := row.Scan(&v.ID, &v.UserID, &v.Email, &v.CodeSalt, &v.CodeHash, &v.ExpiresAt, &v.AttemptsCount, &v.MaxAttempts, &v.Status, &v.VerifiedAt, &v.CreatedAt); err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *Store) GetVerificationByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*ports.EmailVerification, error) {
	row := tx.QueryRow(ctx, `
SELECT id, user_id, email, code_salt, code_hash, expires_at, attempts_count, max_attempts, status, verified_at, created_at
FROM auth_email_verifications WHERE id=$1 FOR UPDATE`, id)
	v, err := scanVerification(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return v, nil
}

func (s *Store) MarkVerificationVerified(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_email_verifications SET status='verified', verified_at=$2 WHERE id=$1 AND status='pending'`, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) IncrementVerificationAttempt(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
UPDATE auth_email_verifications SET attempts_count=attempts_count+1 WHERE id=$1`, id)
	return err
}

func (s *Store) CancelPendingResets(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
UPDATE auth_password_resets SET status='cancelled' WHERE user_id=$1 AND status='pending'`, userID)
	return err
}

func (s *Store) InsertPasswordReset(ctx context.Context, tx pgx.Tx, userID uuid.UUID, email string, salt []byte, codeHash string, expires time.Time) (*ports.PasswordReset, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO auth_password_resets (user_id, email, code_salt, code_hash, expires_at, attempts_count, status)
VALUES ($1,$2,$3,$4,$5,0,'pending')
RETURNING id, user_id, email, code_salt, code_hash, expires_at, attempts_count, status, used_at`, userID, email, salt, codeHash, expires)
	var r ports.PasswordReset
	if err := row.Scan(&r.ID, &r.UserID, &r.Email, &r.CodeSalt, &r.CodeHash, &r.ExpiresAt, &r.AttemptsCount, &r.Status, &r.UsedAt); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) GetPasswordResetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*ports.PasswordReset, error) {
	row := tx.QueryRow(ctx, `
SELECT id, user_id, email, code_salt, code_hash, expires_at, attempts_count, status, used_at
FROM auth_password_resets WHERE id=$1 FOR UPDATE`, id)
	var r ports.PasswordReset
	if err := row.Scan(&r.ID, &r.UserID, &r.Email, &r.CodeSalt, &r.CodeHash, &r.ExpiresAt, &r.AttemptsCount, &r.Status, &r.UsedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *Store) MarkPasswordResetUsed(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_password_resets SET status='used', used_at=$2 WHERE id=$1 AND status='pending'`, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) IncrementPasswordResetAttempt(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `UPDATE auth_password_resets SET attempts_count=attempts_count+1 WHERE id=$1`, id)
	return err
}

func (s *Store) InsertSession(ctx context.Context, tx pgx.Tx, sess ports.Session) error {
	var ip any
	if sess.IP != nil {
		ip = *sess.IP
	}
	_, err := tx.Exec(ctx, `
INSERT INTO auth_sessions (
  id, user_id, refresh_token_hash, session_family_id, replaced_by_session_id,
  device_id, device_name, platform, ip, user_agent, issued_at, expires_at, last_seen_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::inet,$10,$11,$12,$13)`,
		sess.ID, sess.UserID, sess.RefreshTokenHash, sess.SessionFamilyID, sess.ReplacedBySessionID,
		sess.DeviceID, sess.DeviceName, sess.Platform, ip, sess.UserAgent, sess.IssuedAt, sess.ExpiresAt, sess.LastSeenAt,
	)
	return err
}

func (s *Store) GetSessionByRefreshHashForUpdate(ctx context.Context, tx pgx.Tx, hash string) (*ports.Session, error) {
	row := tx.QueryRow(ctx, `
SELECT id, user_id, refresh_token_hash, session_family_id, replaced_by_session_id,
       device_id, device_name, platform, host(ip), user_agent, issued_at, expires_at, last_seen_at,
       revoked_at, revoke_reason, compromised_at
FROM auth_sessions WHERE refresh_token_hash=$1 FOR UPDATE`, hash)
	return scanSession(row)
}

func (s *Store) GetSessionByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*ports.Session, error) {
	row := tx.QueryRow(ctx, `
SELECT id, user_id, refresh_token_hash, session_family_id, replaced_by_session_id,
       device_id, device_name, platform, host(ip), user_agent, issued_at, expires_at, last_seen_at,
       revoked_at, revoke_reason, compromised_at
FROM auth_sessions WHERE id=$1`, id)
	return scanSession(row)
}

func scanSessionFromRows(rows pgx.Rows) (*ports.Session, error) {
	var sess ports.Session
	var ipStr *string
	if err := rows.Scan(
		&sess.ID, &sess.UserID, &sess.RefreshTokenHash, &sess.SessionFamilyID, &sess.ReplacedBySessionID,
		&sess.DeviceID, &sess.DeviceName, &sess.Platform, &ipStr, &sess.UserAgent,
		&sess.IssuedAt, &sess.ExpiresAt, &sess.LastSeenAt, &sess.RevokedAt, &sess.RevokeReason, &sess.CompromisedAt,
	); err != nil {
		return nil, err
	}
	if ipStr != nil && *ipStr != "" {
		sess.IP = ipStr
	}
	return &sess, nil
}

func scanSession(row pgx.Row) (*ports.Session, error) {
	var sess ports.Session
	var ipStr *string
	if err := row.Scan(
		&sess.ID, &sess.UserID, &sess.RefreshTokenHash, &sess.SessionFamilyID, &sess.ReplacedBySessionID,
		&sess.DeviceID, &sess.DeviceName, &sess.Platform, &ipStr, &sess.UserAgent,
		&sess.IssuedAt, &sess.ExpiresAt, &sess.LastSeenAt, &sess.RevokedAt, &sess.RevokeReason, &sess.CompromisedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if ipStr != nil && *ipStr != "" {
		sess.IP = ipStr
	}
	return &sess, nil
}

func (s *Store) ListSessionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]ports.Session, error) {
	rows, err := tx.Query(ctx, `
SELECT id, user_id, refresh_token_hash, session_family_id, replaced_by_session_id,
       device_id, device_name, platform, host(ip), user_agent, issued_at, expires_at, last_seen_at,
       revoked_at, revoke_reason, compromised_at
FROM auth_sessions WHERE user_id=$1 ORDER BY issued_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ports.Session
	for rows.Next() {
		sess, err := scanSessionFromRows(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *sess)
	}
	return list, rows.Err()
}

func (s *Store) RevokeSession(ctx context.Context, tx pgx.Tx, id uuid.UUID, reason string, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_sessions SET revoked_at=$2, revoke_reason=$3 WHERE id=$1 AND revoked_at IS NULL`, id, at, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) RevokeSessionAsRotation(ctx context.Context, tx pgx.Tx, oldID, newID uuid.UUID, at time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE auth_sessions
SET revoked_at=$3, revoke_reason='rotation', replaced_by_session_id=$2
WHERE id=$1 AND revoked_at IS NULL`, oldID, newID, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) RevokeAllSessionsForUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, reason string, at time.Time) (int64, error) {
	tag, err := tx.Exec(ctx, `
UPDATE auth_sessions SET revoked_at=$2, revoke_reason=$3
WHERE user_id=$1 AND revoked_at IS NULL`, userID, at, reason)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (s *Store) RevokeSessionFamily(ctx context.Context, tx pgx.Tx, familyID uuid.UUID, reason string, at time.Time, compromised *time.Time) (int64, error) {
	tag, err := tx.Exec(ctx, `
UPDATE auth_sessions
SET revoked_at=$1, revoke_reason=$2, compromised_at=COALESCE($3, compromised_at)
WHERE session_family_id=$4 AND revoked_at IS NULL`,
		at, reason, compromised, familyID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (s *Store) ListAuditByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, limit, offset int) ([]ports.AuditEvent, error) {
	rows, err := tx.Query(ctx, `
SELECT id, user_id, event_type, host(ip), user_agent, meta, created_at
FROM auth_audit_log WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.AuditEvent
	for rows.Next() {
		var e ports.AuditEvent
		var raw []byte
		var ipStr *string
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventType, &ipStr, &e.UserAgent, &raw, &e.CreatedAt); err != nil {
			return nil, err
		}
		if ipStr != nil && *ipStr != "" {
			e.IP = ipStr
		}
		if err := json.Unmarshal(raw, &e.Meta); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) InsertAudit(ctx context.Context, tx pgx.Tx, e ports.AuditEvent) error {
	meta, err := json.Marshal(e.Meta)
	if err != nil {
		return err
	}
	var ip any
	if e.IP != nil {
		ip = *e.IP
	}
	ct := e.CreatedAt
	if ct.IsZero() {
		ct = time.Now().UTC()
	}
	_, err = tx.Exec(ctx, `
INSERT INTO auth_audit_log (id, user_id, event_type, ip, user_agent, meta, created_at)
VALUES ($1,$2,$3,$4::inet,$5,$6::jsonb, $7)
ON CONFLICT (id, created_at) DO NOTHING`,
		e.ID, e.UserID, e.EventType, ip, e.UserAgent, meta, ct)
	return err
}

func (s *Store) InsertOutbox(ctx context.Context, tx pgx.Tx, m ports.OutboxMessage) error {
	payload, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO auth_outbox (aggregate_type, event_type, payload) VALUES ($1,$2,$3::jsonb)`,
		m.AggregateType, m.EventType, payload)
	return err
}

func (s *Store) ListUnpublishedOutbox(ctx context.Context, tx pgx.Tx, limit int) ([]ports.OutboxRow, error) {
	rows, err := tx.Query(ctx, `
SELECT id, aggregate_type, event_type, payload
FROM auth_outbox WHERE published_at IS NULL ORDER BY id ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.OutboxRow
	for rows.Next() {
		var r ports.OutboxRow
		var raw []byte
		if err := rows.Scan(&r.ID, &r.AggregateType, &r.EventType, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &r.Payload); err != nil {
			return nil, fmt.Errorf("outbox payload: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) MarkOutboxPublished(ctx context.Context, tx pgx.Tx, id int64, at time.Time) error {
	_, err := tx.Exec(ctx, `UPDATE auth_outbox SET published_at=$2 WHERE id=$1`, id, at)
	return err
}
