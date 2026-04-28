package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/jackc/pgx/v5"
)

// Account row projection for auth layer.
// Email/EmailNormalized are nil until the user adds an email in settings.
type Account struct {
	ID              uuid.UUID
	Email           *string
	EmailNormalized *string
	PhoneE164       *string
	PhoneVerifiedAt *time.Time
	Status          domain.AccountStatus
	Role            domain.Role
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	BlockedAt       *time.Time
	BlockedReason   *string
	DeletedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Credential row.
type Credential struct {
	UserID            uuid.UUID
	PasswordHash      string
	AlgoVersion       int16
	PasswordChangedAt time.Time
	FailedLoginCount  int
	LastFailedLoginAt *time.Time
}

// EmailVerification row.
type EmailVerification struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Email         string
	CodeSalt      []byte
	CodeHash      string
	ExpiresAt     time.Time
	AttemptsCount int
	MaxAttempts   int
	Status        string
	VerifiedAt    *time.Time
	CreatedAt     time.Time
}

// PasswordReset row.
type PasswordReset struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Email         string
	CodeSalt      []byte
	CodeHash      string
	ExpiresAt     time.Time
	AttemptsCount int
	Status        string
	UsedAt        *time.Time
}

// Session row.
type Session struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	RefreshTokenHash    string
	SessionFamilyID     uuid.UUID
	ReplacedBySessionID *uuid.UUID
	DeviceID            *string
	DeviceName          *string
	Platform            *string
	IP                  *string
	UserAgent           *string
	IssuedAt            time.Time
	ExpiresAt           time.Time
	LastSeenAt          *time.Time
	RevokedAt           *time.Time
	RevokeReason        *string
	CompromisedAt       *time.Time
}

// AuditEvent for persistence.
type AuditEvent struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	EventType string
	IP        *string
	UserAgent *string
	Meta      map[string]any
	CreatedAt time.Time
}

// OutboxMessage for transactional outbox.
type OutboxMessage struct {
	AggregateType string
	EventType     string
	Payload       map[string]any
}

// OutboxRow is an unpublished outbox entry.
type OutboxRow struct {
	ID            int64
	AggregateType string
	EventType     string
	Payload       map[string]any
}

// ProfileReservation reserves nickname in Profile service.
type ProfileReservation struct {
	ReservationID uuid.UUID
}

// ProfileClient abstracts Profile service.
type ProfileClient interface {
	ReserveNickname(ctx context.Context, nickname string) (ProfileReservation, error)
	ReleaseReservation(ctx context.Context, reservationID uuid.UUID) error
	ConfirmReservation(ctx context.Context, reservationID uuid.UUID, userID uuid.UUID) error
}

// EmailSender sends transactional email (OTP, etc.).
type EmailSender interface {
	SendVerificationOTP(ctx context.Context, toEmail, code string) error
	SendPasswordResetOTP(ctx context.Context, toEmail, code string) error
}

// SMSSender sends transactional SMS (OTP).
type SMSSender interface {
	SendOTP(ctx context.Context, phoneE164, code string) error
}

// TokenPair is issued access + refresh.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessExp    time.Time
}

// JWTSigner issues and parses access tokens.
type JWTSigner interface {
	SignAccess(ctx context.Context, userID, sessionID uuid.UUID, role domain.Role, ttl time.Duration) (token string, jti string, exp time.Time, err error)
	ParseAccess(ctx context.Context, token string) (claims AccessClaims, err error)
	JWKS(ctx context.Context) (json []byte, err error)
}

// AccessClaims extracted from JWT.
type AccessClaims struct {
	Sub      uuid.UUID
	Session  uuid.UUID
	Role     domain.Role
	JTI      string
	Expires  time.Time
	IssuedAt time.Time
}

// RateLimiter enforces sliding-window limits (fail-closed on Redis errors).
type RateLimiter interface {
	Allow(ctx context.Context, key string, window time.Duration, limit int) (allowed bool, remaining int, reset time.Time, err error)
}

// Lockout tracks failed login lockouts in Redis.
type Lockout interface {
	IsLocked(ctx context.Context, userID uuid.UUID) (bool, time.Duration, error)
	RecordFailure(ctx context.Context, userID uuid.UUID) (locked bool, ttl time.Duration, err error)
	Clear(ctx context.Context, userID uuid.UUID) error
}

// SessionCache stores short-lived grace tokens after refresh rotation.
type SessionCache interface {
	SetRefreshGrace(ctx context.Context, oldRefreshHash string, pair TokenPair, ttl time.Duration) error
	GetRefreshGrace(ctx context.Context, oldRefreshHash string) (*TokenPair, error)
}

// AccessBlacklist invalidates access tokens by jti until expiry.
type AccessBlacklist interface {
	Add(ctx context.Context, jti string, ttl time.Duration) error
	Has(ctx context.Context, jti string) (bool, error)
}

// PhoneOTPVerifiedPayload is returned after a successful phone OTP challenge consumption.
type PhoneOTPVerifiedPayload struct {
	UserID    uuid.UUID
	PhoneE164 string
}

// PhoneOTPChallengeStore holds short-lived SMS OTP challenges in Redis.
type PhoneOTPChallengeStore interface {
	// CancelPendingAndCreate cancels all pending phone OTPs for the user and stores a new challenge.
	CancelPendingAndCreate(ctx context.Context, userID uuid.UUID, phoneE164, purpose string, salt []byte, codeHash string, ttl time.Duration) (verificationID uuid.UUID, err error)
	// VerifyAndConsume checks the code, increments attempts on mismatch, and removes the challenge on success.
	VerifyAndConsume(ctx context.Context, verificationID uuid.UUID, purpose, plainCode string) (*PhoneOTPVerifiedPayload, error)
}

// Store is the persistence port for auth (Postgres).
type Store interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error

	GetAccountByEmailNorm(ctx context.Context, tx pgx.Tx, emailNorm string) (*Account, error)
	GetAccountByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Account, error)
	GetAccountByPhoneE164(ctx context.Context, tx pgx.Tx, phoneE164 string) (*Account, error)
	InsertAccount(ctx context.Context, tx pgx.Tx, email, emailNorm string, status domain.AccountStatus, role domain.Role) (*Account, error)
	InsertAccountPhone(ctx context.Context, tx pgx.Tx, phoneE164 string, status domain.AccountStatus, role domain.Role) (*Account, error)
	MarkAccountPhoneVerified(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error
	SetAccountPendingEmail(ctx context.Context, tx pgx.Tx, id uuid.UUID, email, emailNorm string) error
	UpdateAccountStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status domain.AccountStatus) error
	SetEmailVerified(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error
	SetBlocked(ctx context.Context, tx pgx.Tx, id uuid.UUID, reason string, at time.Time) error
	ClearBlocked(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	SoftDeleteAccount(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error
	UpdateLastLogin(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error

	UpsertCredential(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string, algo int16) error
	GetCredential(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*Credential, error)
	UpdateCredentialFailures(ctx context.Context, tx pgx.Tx, userID uuid.UUID, count int, lastFail *time.Time) error
	ResetCredentialFailures(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	UpdatePasswordHash(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string, algo int16, changedAt time.Time) error

	CancelPendingVerifications(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	InsertVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, email string, salt []byte, codeHash string, expires time.Time) (*EmailVerification, error)
	GetVerificationByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*EmailVerification, error)
	MarkVerificationVerified(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error
	IncrementVerificationAttempt(ctx context.Context, tx pgx.Tx, id uuid.UUID) error

	CancelPendingResets(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	InsertPasswordReset(ctx context.Context, tx pgx.Tx, userID uuid.UUID, email string, salt []byte, codeHash string, expires time.Time) (*PasswordReset, error)
	GetPasswordResetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*PasswordReset, error)
	MarkPasswordResetUsed(ctx context.Context, tx pgx.Tx, id uuid.UUID, at time.Time) error
	IncrementPasswordResetAttempt(ctx context.Context, tx pgx.Tx, id uuid.UUID) error

	InsertSession(ctx context.Context, tx pgx.Tx, s Session) error
	GetSessionByRefreshHashForUpdate(ctx context.Context, tx pgx.Tx, hash string) (*Session, error)
	GetSessionByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Session, error)
	ListSessionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]Session, error)
	RevokeSession(ctx context.Context, tx pgx.Tx, id uuid.UUID, reason string, at time.Time) error
	RevokeSessionAsRotation(ctx context.Context, tx pgx.Tx, oldID, newID uuid.UUID, at time.Time) error
	RevokeAllSessionsForUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, reason string, at time.Time) (int64, error)
	RevokeSessionFamily(ctx context.Context, tx pgx.Tx, familyID uuid.UUID, reason string, at time.Time, compromised *time.Time) (int64, error)

	InsertAudit(ctx context.Context, tx pgx.Tx, e AuditEvent) error
	ListAuditByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, limit, offset int) ([]AuditEvent, error)
	InsertOutbox(ctx context.Context, tx pgx.Tx, m OutboxMessage) error
	ListUnpublishedOutbox(ctx context.Context, tx pgx.Tx, limit int) ([]OutboxRow, error)
	MarkOutboxPublished(ctx context.Context, tx pgx.Tx, id int64, at time.Time) error
}
