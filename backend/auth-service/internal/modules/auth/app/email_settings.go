package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// StartEmailVerification sets pending email and sends OTP (authenticated user in settings).
// Returns the verification row id for POST /v1/auth/verify-email.
func (s *Service) StartEmailVerification(ctx context.Context, userID uuid.UUID, email string) (uuid.UUID, error) {
	emailAddr, norm, err := normalizeEmail(email)
	if err != nil {
		return uuid.Nil, err
	}
	var verID uuid.UUID
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.Store.GetAccountByID(ctx, tx, userID)
		if err != nil {
			return err
		}
		if acc.Status != domain.StatusActive {
			return domain.ErrForbidden
		}
		if acc.EmailVerifiedAt != nil {
			return domain.ErrConflict
		}
		if o, e2 := s.Store.GetAccountByEmailNorm(ctx, tx, norm); e2 == nil && o != nil && o.ID != userID {
			return domain.ErrConflict
		} else if e2 != nil && !errors.Is(e2, domain.ErrNotFound) {
			return e2
		}
		if err := s.Store.SetAccountPendingEmail(ctx, tx, userID, emailAddr, norm); err != nil {
			return err
		}
		if err := s.Store.CancelPendingVerifications(ctx, tx, userID); err != nil {
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
		v, err := s.Store.InsertVerification(ctx, tx, userID, emailAddr, salt, codeHash, exp)
		if err != nil {
			return err
		}
		verID = v.ID
		_ = s.Mail.SendVerificationOTP(ctx, emailAddr, plain)
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &userID,
			EventType: "email_verification_started",
			Meta:      map[string]any{"email_domain": domainFromEmail(norm)},
			CreatedAt: time.Now().UTC(),
		})
	})
	if err != nil {
		return uuid.Nil, err
	}
	return verID, nil
}

// ResendEmailVerification resends the OTP for the current user's pending email proof.
// Returns the new verification id for POST /v1/auth/verify-email.
func (s *Service) ResendEmailVerification(ctx context.Context, userID uuid.UUID) (RateLimitInfo, uuid.UUID, error) {
	var zero RateLimitInfo
	var zeroID uuid.UUID
	acc, err := func() (*ports.Account, error) {
		var out *ports.Account
		err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			a, err := s.Store.GetAccountByID(ctx, tx, userID)
			if err != nil {
				return err
			}
			out = a
			return nil
		})
		return out, err
	}()
	if err != nil {
		return zero, zeroID, err
	}
	if acc == nil {
		return zero, zeroID, domain.ErrNotFound
	}
	if acc.Status != domain.StatusActive {
		return zero, zeroID, domain.ErrForbidden
	}
	if acc.Email == nil || *acc.Email == "" {
		return zero, zeroID, domain.ErrBadRequest
	}
	if acc.EmailVerifiedAt != nil {
		return zero, zeroID, domain.ErrConflict
	}
	ok, rl, err := s.rateConsume(ctx, "rl:resend:email:uid:"+userID.String(), time.Minute, 5)
	if err != nil {
		return rl, zeroID, err
	}
	if !ok {
		return rl, zeroID, domain.ErrRateLimited
	}
	okCd, _, err := s.Redis.ResendCooldown(ctx, acc.ID, resendCooldown)
	if err != nil {
		return rl, zeroID, domain.ErrServiceUnavailable
	}
	if !okCd {
		return rl, zeroID, domain.ErrRateLimited
	}
	emailTo := *acc.Email
	salt, err := crypto.RandomSalt(16)
	if err != nil {
		return rl, zeroID, err
	}
	plain, err := crypto.RandomDigits(6)
	if err != nil {
		return rl, zeroID, err
	}
	hash := crypto.HashOTP(salt, plain)
	exp := time.Now().Add(s.VerifyTTL)
	var verID uuid.UUID
	if err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.Store.CancelPendingVerifications(ctx, tx, acc.ID); err != nil {
			return err
		}
		v, err := s.Store.InsertVerification(ctx, tx, acc.ID, emailTo, salt, hash, exp)
		if err != nil {
			return err
		}
		verID = v.ID
		return nil
	}); err != nil {
		return rl, zeroID, err
	}
	_ = s.Mail.SendVerificationOTP(ctx, emailTo, plain)
	return rl, verID, nil
}

// SetInitialPassword sets the first password (no prior credential row). Optional after SMS login.
func (s *Service) SetInitialPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	hash, err := s.Hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := s.Store.GetCredential(ctx, tx, userID)
		if err == nil {
			return domain.ErrConflict
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		now := time.Now().UTC()
		if err := s.Store.UpsertCredential(ctx, tx, userID, hash, 1); err != nil {
			return err
		}
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &userID,
			EventType: "password_set_initial",
			Meta:      map[string]any{},
			CreatedAt: now,
		})
	})
}
