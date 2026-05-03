package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// RegisterEmail creates a new account with email+password and sends a verification OTP.
// Returns the verification row ID and the OTP plain code (for demo / dev mode).
func (s *Service) RegisterEmail(ctx context.Context, email, password, name string) (uuid.UUID, string, error) {
	if name == "" {
		return uuid.Nil, "", domain.ErrBadRequest
	}
	addr, norm, err := normalizeEmail(email)
	if err != nil {
		return uuid.Nil, "", err
	}

	var verID uuid.UUID
	var plainCode string
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		existing, _ := s.Store.GetAccountByEmailNorm(ctx, tx, norm)
		if existing != nil && existing.Status != domain.StatusDeleted {
			return domain.ErrConflict
		}

		acc, err := s.Store.InsertAccount(ctx, tx, addr, norm,
			domain.StatusPendingVerification, domain.RoleUser)
		if err != nil {
			return err
		}

		hash, err := s.Hasher.Hash(password)
		if err != nil {
			return err
		}
		if err := s.Store.UpsertCredential(ctx, tx, acc.ID, hash, 1); err != nil {
			return err
		}

		salt, err := crypto.RandomSalt(16)
		if err != nil {
			return err
		}
		plain, err := crypto.RandomDigits(6)
		if err != nil {
			return err
		}
		codeHash := crypto.HashOTP(salt, plain)
		exp := time.Now().Add(s.VerifyTTL)

		v, err := s.Store.InsertVerification(ctx, tx, acc.ID, addr, salt, codeHash, exp)
		if err != nil {
			return err
		}

		verID = v.ID
		plainCode = plain
		go s.Mail.SendVerificationOTP(context.Background(), addr, plain)

		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &acc.ID,
			EventType: "email_registration_started",
			Meta:      map[string]any{"email_domain": domainFromEmail(norm)},
			CreatedAt: time.Now().UTC(),
		})
	})
	return verID, plainCode, err
}
