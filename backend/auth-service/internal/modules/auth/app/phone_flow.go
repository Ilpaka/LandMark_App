package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/metrics"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/phone"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

const defaultPhoneRegion = "RU"

// PhoneRequestInput is the first step: send SMS to this raw phone string (international or local).
type PhoneRequestInput struct {
	Phone   string
	IP      *string
	Default string // region for parsing, e.g. RU; if empty, defaultPhoneRegion
}

// PhoneRequestResult returns the OTP record id the client must send back to verify.
type PhoneRequestResult struct {
	VerificationID uuid.UUID
	Rate           RateLimitInfo
}

// RequestPhoneCode creates a pending account on first use or issues login OTP for an existing one.
func (s *Service) RequestPhoneCode(ctx context.Context, in PhoneRequestInput) (*PhoneRequestResult, error) {
	region := in.Default
	if region == "" {
		region = defaultPhoneRegion
	}
	e164, err := phone.ToE164(in.Phone, region)
	if err != nil {
		return nil, domain.ErrBadRequest
	}
	ip := deref(in.IP)
	if ip == "" {
		ip = "unknown"
	}
	okIP, infoIP, err := s.rateConsume(ctx, "rl:phone:ip:"+ip, rateRegisterWindow, rateRegisterLimit)
	if err != nil {
		return nil, err
	}
	if !okIP {
		return &PhoneRequestResult{Rate: infoIP}, domain.ErrRateLimited
	}
	okPh, infoPh, err := s.rateConsume(ctx, "rl:phone:e164:"+e164, time.Minute, 5)
	if err != nil {
		return nil, err
	}
	if !okPh {
		rl := mergeRateLimit(infoIP, infoPh)
		return &PhoneRequestResult{Rate: rl}, domain.ErrRateLimited
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

	var outID uuid.UUID
	var newUser bool
	var userID uuid.UUID
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, aerr := s.Store.GetAccountByPhoneE164(ctx, tx, e164)
		if aerr != nil && !errors.Is(aerr, domain.ErrNotFound) {
			return aerr
		}
		if errors.Is(aerr, domain.ErrNotFound) {
			ins, err := s.Store.InsertAccountPhone(ctx, tx, e164, domain.StatusPendingVerification, domain.RoleUser)
			if err != nil {
				if isPGUnique(err) {
					return domain.ErrConflict
				}
				return err
			}
			acc = ins
			userID = acc.ID
			newUser = true
			if err := s.Store.InsertOutbox(ctx, tx, ports.OutboxMessage{
				AggregateType: "user",
				EventType:     "user.registered",
				Payload: map[string]any{
					"user_id":     acc.ID.String(),
					"phone_e164":  e164,
					"pending_sms": true,
				},
			}); err != nil {
				return err
			}
		} else {
			if acc == nil {
				return domain.ErrServiceUnavailable
			}
			userID = acc.ID
			switch acc.Status {
			case domain.StatusBlocked:
				return domain.ErrAccountBlocked
			case domain.StatusDeleted:
				return domain.ErrAccountDeleted
			}
		}
		sfx := e164
		if len(sfx) > 4 {
			sfx = sfx[len(sfx)-4:]
		}
		if newUser {
			_ = s.enqueueAudit(ctx, tx, ports.AuditEvent{
				ID:        uuid.New(),
				UserID:    &userID,
				EventType: "phone_register_started",
				Meta:      map[string]any{"phone_suffix": sfx},
				CreatedAt: time.Now().UTC(),
			})
		} else {
			_ = s.enqueueAudit(ctx, tx, ports.AuditEvent{
				ID:        uuid.New(),
				UserID:    &userID,
				EventType: "phone_login_code_sent",
				Meta:      map[string]any{"phone_suffix": sfx},
				CreatedAt: time.Now().UTC(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	outID, err = s.PhoneOTP.CancelPendingAndCreate(ctx, userID, e164, domain.PhoneOTPPurposeLogin, salt, codeHash, s.VerifyTTL)
	if err != nil {
		metrics.AuthPhoneOTPStoreErrors.Inc()
		return nil, domain.ErrServiceUnavailable
	}
	if newUser {
		metrics.AuthRegisterTotal.Inc()
	}
	if err := s.SMS.SendOTP(ctx, e164, plain); err != nil {
		s.Log.Warn("send sms", "err", err)
	}
	return &PhoneRequestResult{VerificationID: outID, Rate: rlOut}, nil
}

func isPGUnique(err error) bool {
	var pe *pgconn.PgError
	return errors.As(err, &pe) && pe.Code == "23505"
}

func (s *Service) notePhoneOTPVerify(err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, domain.ErrInvalidOTP):
		metrics.AuthPhoneOTPVerifyOutcomes.WithLabelValues("invalid_code").Inc()
	case errors.Is(err, domain.ErrTooManyAttempts):
		metrics.AuthPhoneOTPVerifyOutcomes.WithLabelValues("exhausted").Inc()
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrConflict):
		metrics.AuthPhoneOTPVerifyOutcomes.WithLabelValues("not_found_or_conflict").Inc()
	default:
		metrics.AuthPhoneOTPVerifyOutcomes.WithLabelValues("other").Inc()
		metrics.AuthPhoneOTPStoreErrors.Inc()
	}
}

// PhoneVerifyInput checks the SMS code and returns tokens.
type PhoneVerifyInput struct {
	VerificationID uuid.UUID
	Code           string
	DeviceID       *string
	DeviceName     *string
	Platform       *string
	IP             *string
	UserAgent      *string
}

// VerifyPhoneCode validates the OTP, activates the user on first success, and issues access + refresh.
func (s *Service) VerifyPhoneCode(ctx context.Context, in PhoneVerifyInput) (*ports.TokenPair, RateLimitInfo, error) {
	var zeroRL RateLimitInfo
	okIP, infoIP, err := s.rateConsume(ctx, "rl:phone:verify:ip:"+deref(in.IP), rateLoginWindow, rateLoginLimit)
	if err != nil {
		return nil, zeroRL, err
	}
	if !okIP {
		return nil, infoIP, domain.ErrRateLimited
	}
	// Limit attempts per verification id bucket (avoid brute force on code)
	okV, infoV, err := s.rateConsume(ctx, "rl:phone:verify:v:"+in.VerificationID.String(), time.Minute, 30)
	if err != nil {
		return nil, zeroRL, err
	}
	if !okV {
		rl := mergeRateLimit(infoIP, infoV)
		return nil, rl, domain.ErrRateLimited
	}
	rlOut := mergeRateLimit(infoIP, infoV)

	payload, err := s.PhoneOTP.VerifyAndConsume(ctx, in.VerificationID, domain.PhoneOTPPurposeLogin, in.Code)
	if err != nil {
		s.notePhoneOTPVerify(err)
		return nil, zeroRL, err
	}
	metrics.AuthPhoneOTPVerifyOutcomes.WithLabelValues("success").Inc()

	var pair *ports.TokenPair
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.Store.GetAccountByID(ctx, tx, payload.UserID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if acc.Status == domain.StatusPendingVerification {
			if err := s.Store.MarkAccountPhoneVerified(ctx, tx, acc.ID, now); err != nil {
				return err
			}
		}
		acc, err = s.Store.GetAccountByID(ctx, tx, payload.UserID)
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
			return domain.ErrInvalidCredentials
		}
		locked, _, lerr := s.Redis.IsLocked(ctx, acc.ID)
		if lerr != nil {
			return domain.ErrServiceUnavailable
		}
		if locked {
			return domain.ErrInvalidCredentials
		}
		var sid uuid.UUID
		pair, sid, err = s.issueFreshSession(ctx, tx, acc, sessionDeviceInfo{
			DeviceID: in.DeviceID, DeviceName: in.DeviceName, Platform: in.Platform,
			IP: in.IP, UserAgent: in.UserAgent,
		}, now)
		if err != nil {
			return err
		}
		_ = s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &acc.ID,
			EventType: "phone_verified",
			IP:        in.IP,
			UserAgent: in.UserAgent,
			Meta:      map[string]any{"session_id": sid.String()},
			CreatedAt: now,
		})
		return nil
	})
	if err != nil {
		return nil, rlOut, err
	}
	if pair != nil {
		metrics.AuthLoginTotal.Inc()
	}
	return pair, rlOut, nil
}

// ResendPhoneCode sends a new code for an existing in-flight flow (phone must match a known account / pending row).
// Body uses raw phone; rate limits apply.
func (s *Service) ResendPhoneCode(ctx context.Context, in PhoneRequestInput) (RateLimitInfo, error) {
	region := in.Default
	if region == "" {
		region = defaultPhoneRegion
	}
	e164, err := phone.ToE164(in.Phone, region)
	if err != nil {
		return RateLimitInfo{}, domain.ErrBadRequest
	}
	rl, err := s.resendPhoneInternal(ctx, e164, in.IP)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return RateLimitInfo{}, nil
		}
		return RateLimitInfo{}, err
	}
	return rl, nil
}

func (s *Service) resendPhoneInternal(ctx context.Context, e164 string, ip *string) (RateLimitInfo, error) {
	ipStr := deref(ip)
	if ipStr == "" {
		ipStr = "unknown"
	}
	ok, rl, err := s.rateConsume(ctx, "rl:phone:resend:ip:"+ipStr, time.Minute, 5)
	if err != nil {
		return rl, err
	}
	if !ok {
		return rl, domain.ErrRateLimited
	}
	return s.resendAfterRate(ctx, e164, rl)
}

func (s *Service) resendAfterRate(ctx context.Context, e164 string, baseRL RateLimitInfo) (RateLimitInfo, error) {
	acc, err := func() (*ports.Account, error) {
		var out *ports.Account
		err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			a, err := s.Store.GetAccountByPhoneE164(ctx, tx, e164)
			if err != nil {
				return err
			}
			out = a
			return nil
		})
		return out, err
	}()
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return baseRL, err
		}
		return baseRL, err
	}
	if acc == nil {
		return baseRL, domain.ErrNotFound
	}
	if acc.Status != domain.StatusActive && acc.Status != domain.StatusPendingVerification {
		return baseRL, nil
	}
	okCd, _, err := s.Redis.ResendCooldown(ctx, acc.ID, resendCooldown)
	if err != nil {
		return baseRL, domain.ErrServiceUnavailable
	}
	if !okCd {
		return baseRL, domain.ErrRateLimited
	}
	salt, err := crypto.RandomSalt(16)
	if err != nil {
		return baseRL, err
	}
	plain, err := crypto.RandomDigits(6)
	if err != nil {
		return baseRL, err
	}
	hash := crypto.HashOTP(salt, plain)
	_, err = s.PhoneOTP.CancelPendingAndCreate(ctx, acc.ID, e164, domain.PhoneOTPPurposeLogin, salt, hash, s.VerifyTTL)
	if err != nil {
		metrics.AuthPhoneOTPStoreErrors.Inc()
		return baseRL, domain.ErrServiceUnavailable
	}
	_ = s.SMS.SendOTP(ctx, e164, plain)
	return baseRL, nil
}
