package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/metrics"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/phone"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// ForgotPasswordPhoneResult is returned from ForgotPasswordPhone (verification_id only when SMS was sent).
type ForgotPasswordPhoneResult struct {
	VerificationID *uuid.UUID
	Rate           RateLimitInfo
}

// ForgotPasswordPhone queues an SMS OTP for password reset when the account is eligible.
// Uses forgotMinLatency like email forgot to reduce timing leaks.
func (s *Service) ForgotPasswordPhone(ctx context.Context, in PhoneRequestInput) (*ForgotPasswordPhoneResult, error) {
	start := time.Now()
	region := in.Default
	if region == "" {
		region = defaultPhoneRegion
	}
	e164, err := phone.ToE164(in.Phone, region)
	if err != nil {
		if d := time.Since(start); d < forgotMinLatency {
			time.Sleep(forgotMinLatency - d)
		}
		return &ForgotPasswordPhoneResult{}, nil
	}
	ip := deref(in.IP)
	if ip == "" {
		ip = "unknown"
	}
	okIP, infoIP, err := s.rateConsume(ctx, "rl:forgot:phone:ip:"+ip, rateRegisterWindow, rateRegisterLimit)
	if err != nil {
		return nil, err
	}
	if !okIP {
		return &ForgotPasswordPhoneResult{Rate: infoIP}, domain.ErrRateLimited
	}
	okPh, infoPh, err := s.rateConsume(ctx, "rl:forgot:phone:e164:"+e164, time.Minute, 5)
	if err != nil {
		return nil, err
	}
	if !okPh {
		rl := mergeRateLimit(infoIP, infoPh)
		return &ForgotPasswordPhoneResult{Rate: rl}, domain.ErrRateLimited
	}
	rlOut := mergeRateLimit(infoIP, infoPh)

	salt, err := crypto.RandomSalt(16)
	if err != nil {
		return nil, err
	}
	plain, err := crypto.RandomDigits(6)
	if err != nil {
		return nil, err
	}
	codeHash := crypto.HashOTP(salt, plain)

	var outID *uuid.UUID
	var eligible *ports.Account
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, aerr := s.Store.GetAccountByPhoneE164(ctx, tx, e164)
		if aerr != nil {
			if errors.Is(aerr, domain.ErrNotFound) {
				return nil
			}
			return aerr
		}
		if acc == nil {
			return nil
		}
		switch acc.Status {
		case domain.StatusBlocked, domain.StatusDeleted:
			return nil
		}
		if acc.Status != domain.StatusActive {
			return nil
		}
		if acc.PhoneVerifiedAt == nil {
			return nil
		}
		if _, cerr := s.Store.GetCredential(ctx, tx, acc.ID); cerr != nil {
			if errors.Is(cerr, domain.ErrNotFound) {
				return nil
			}
			return cerr
		}
		eligible = acc
		sfx := e164
		if len(sfx) > 4 {
			sfx = sfx[len(sfx)-4:]
		}
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &acc.ID,
			EventType: "phone_password_reset_started",
			IP:        in.IP,
			Meta:      map[string]any{"phone_suffix": sfx},
			CreatedAt: time.Now().UTC(),
		})
	})
	if err != nil {
		return nil, err
	}
	if eligible != nil {
		otpID, perr := s.PhoneOTP.CancelPendingAndCreate(ctx, eligible.ID, e164, domain.PhoneOTPPurposePasswordReset, salt, codeHash, s.VerifyTTL)
		if perr != nil {
			metrics.AuthPhoneOTPStoreErrors.Inc()
			return nil, domain.ErrServiceUnavailable
		}
		outID = &otpID
		_ = s.SMS.SendOTP(ctx, e164, plain)
	}
	if d := time.Since(start); d < forgotMinLatency {
		time.Sleep(forgotMinLatency - d)
	}
	return &ForgotPasswordPhoneResult{VerificationID: outID, Rate: rlOut}, nil
}

// ResetPasswordPhoneInput confirms SMS code and sets a new password (no tokens).
type ResetPasswordPhoneInput struct {
	VerificationID uuid.UUID
	Code           string
	NewPassword    string
	IP             *string
	UserAgent      *string
}

// ResetPasswordPhone validates a password_reset phone OTP and updates the password.
func (s *Service) ResetPasswordPhone(ctx context.Context, in ResetPasswordPhoneInput) (RateLimitInfo, error) {
	var zeroRL RateLimitInfo
	okIP, infoIP, err := s.rateConsume(ctx, "rl:phone:reset:ip:"+deref(in.IP), rateLoginWindow, rateLoginLimit)
	if err != nil {
		return zeroRL, err
	}
	if !okIP {
		return infoIP, domain.ErrRateLimited
	}
	okV, infoV, err := s.rateConsume(ctx, "rl:phone:reset:v:"+in.VerificationID.String(), time.Minute, 30)
	if err != nil {
		return zeroRL, err
	}
	if !okV {
		return mergeRateLimit(infoIP, infoV), domain.ErrRateLimited
	}
	rlOut := mergeRateLimit(infoIP, infoV)

	pwdHash, err := s.Hasher.Hash(in.NewPassword)
	if err != nil {
		return rlOut, err
	}

	payload, err := s.PhoneOTP.VerifyAndConsume(ctx, in.VerificationID, domain.PhoneOTPPurposePasswordReset, in.Code)
	if err != nil {
		s.notePhoneOTPVerify(err)
		return zeroRL, err
	}
	metrics.AuthPhoneOTPVerifyOutcomes.WithLabelValues("success").Inc()

	var userID uuid.UUID
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.Store.GetAccountByID(ctx, tx, payload.UserID)
		if err != nil {
			return err
		}
		switch acc.Status {
		case domain.StatusBlocked:
			return domain.ErrAccountBlocked
		case domain.StatusDeleted:
			return domain.ErrAccountDeleted
		}
		if acc.Status != domain.StatusActive {
			return domain.ErrInvalidOTP
		}
		if _, err := s.Store.GetCredential(ctx, tx, acc.ID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrInvalidOTP
			}
			return err
		}
		now := time.Now().UTC()
		if err := s.Store.UpdatePasswordHash(ctx, tx, acc.ID, pwdHash, 1, now); err != nil {
			return err
		}
		if _, err := s.Store.RevokeAllSessionsForUser(ctx, tx, acc.ID, "password_reset", now); err != nil {
			return err
		}
		userID = acc.ID
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &acc.ID,
			EventType: "phone_password_reset_completed",
			IP:        in.IP,
			UserAgent: in.UserAgent,
			Meta:      map[string]any{},
			CreatedAt: now,
		})
	})
	if err != nil {
		return rlOut, err
	}
	_ = s.Redis.Clear(ctx, userID)
	return rlOut, nil
}
