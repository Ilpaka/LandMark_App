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
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

const (
	graceRefresh       = 10 * time.Second
	resendCooldown     = 60 * time.Second
	forgotMinLatency   = 200 * time.Millisecond
	rateRegisterLimit  = 10
	rateRegisterWindow = time.Minute
	rateLoginLimit     = 30
	rateLoginWindow    = time.Minute
)

func mergeStr(old *string, neu *string) *string {
	if neu != nil && *neu != "" {
		return neu
	}
	return old
}

func (s *Service) VerifyEmail(ctx context.Context, verificationID uuid.UUID, code string) error {
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		v, err := s.Store.GetVerificationByIDForUpdate(ctx, tx, verificationID)
		if err != nil {
			return err
		}
		if v.Status != "pending" {
			return domain.ErrConflict
		}
		if time.Now().After(v.ExpiresAt) {
			return domain.ErrInvalidOTP
		}
		if v.AttemptsCount >= v.MaxAttempts {
			return domain.ErrTooManyAttempts
		}
		if crypto.HashOTP(v.CodeSalt, code) != v.CodeHash {
			_ = s.Store.IncrementVerificationAttempt(ctx, tx, v.ID)
			return domain.ErrInvalidOTP
		}
		now := time.Now().UTC()
		if err := s.Store.MarkVerificationVerified(ctx, tx, v.ID, now); err != nil {
			return err
		}
		if err := s.Store.SetEmailVerified(ctx, tx, v.UserID, now); err != nil {
			return err
		}
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &v.UserID,
			EventType: "email_verified",
			Meta:      map[string]any{},
			CreatedAt: now,
		})
	})
}

func deref(s *string) string {
	if s == nil {
		return "unknown"
	}
	return *s
}

// LoginPasswordInput is email/password login with optional device metadata.
type LoginPasswordInput struct {
	Email      string
	Password   string
	DeviceID   *string
	DeviceName *string
	Platform   *string
	IP         *string
	UserAgent  *string
}

// LoginWithPassword issues tokens when email is verified and password matches.
func (s *Service) LoginWithPassword(ctx context.Context, in LoginPasswordInput) (*ports.TokenPair, RateLimitInfo, error) {
	var zeroRL RateLimitInfo
	okIP, infoIP, err := s.rateConsume(ctx, "rl:login:email:ip:"+deref(in.IP), rateLoginWindow, rateLoginLimit)
	if err != nil {
		return nil, zeroRL, err
	}
	if !okIP {
		return nil, infoIP, domain.ErrRateLimited
	}
	_, norm, err := normalizeEmail(in.Email)
	if err != nil {
		return nil, infoIP, err
	}

	var pair *ports.TokenPair
	var recordFail *uuid.UUID
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.Store.GetAccountByEmailNorm(ctx, tx, norm)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrInvalidCredentials
			}
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
		if acc.EmailVerifiedAt == nil {
			return domain.ErrEmailNotVerified
		}
		cred, err := s.Store.GetCredential(ctx, tx, acc.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrInvalidCredentials
			}
			return err
		}
		locked, _, lerr := s.Redis.IsLocked(ctx, acc.ID)
		if lerr != nil {
			return domain.ErrServiceUnavailable
		}
		if locked {
			return domain.ErrInvalidCredentials
		}
		okPwd, verr := s.Hasher.Verify(cred.PasswordHash, in.Password)
		if verr != nil || !okPwd {
			uid := acc.ID
			recordFail = &uid
			return domain.ErrInvalidCredentials
		}
		if err := s.Store.ResetCredentialFailures(ctx, tx, acc.ID); err != nil {
			return err
		}
		now := time.Now().UTC()
		var sid uuid.UUID
		pair, sid, err = s.issueFreshSession(ctx, tx, acc, sessionDeviceInfo{
			DeviceID: in.DeviceID, DeviceName: in.DeviceName, Platform: in.Platform,
			IP: in.IP, UserAgent: in.UserAgent,
		}, now)
		if err != nil {
			return err
		}
		if err := s.Redis.Clear(ctx, acc.ID); err != nil {
			return err
		}
		_ = s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &acc.ID,
			EventType: "password_login",
			IP:        in.IP,
			UserAgent: in.UserAgent,
			Meta:      map[string]any{"session_id": sid.String()},
			CreatedAt: now,
		})
		return nil
	})
	if err != nil && recordFail != nil {
		_, _, _ = s.Redis.RecordFailure(ctx, *recordFail)
	}
	if err != nil {
		return nil, infoIP, err
	}
	if pair != nil {
		metrics.AuthLoginTotal.Inc()
	}
	return pair, infoIP, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, deviceID *string, ip *string, ua *string) (*ports.TokenPair, error) {
	oldHash := crypto.RefreshTokenHash(s.Pepper, refreshToken)
	if pair, err := s.Redis.GetRefreshGrace(ctx, oldHash); err != nil {
		return nil, err
	} else if pair != nil {
		return pair, nil
	}
	var out *ports.TokenPair
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		sess, err := s.Store.GetSessionByRefreshHashForUpdate(ctx, tx, oldHash)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrInvalidRefresh
			}
			return err
		}
		now := time.Now().UTC()
		if sess.RevokedAt != nil {
			if sess.RevokeReason != nil && *sess.RevokeReason == "rotation" {
				cmp := now
				if _, rerr := s.Store.RevokeSessionFamily(ctx, tx, sess.SessionFamilyID, "refresh_reuse_detected", now, &cmp); rerr != nil {
					return rerr
				}
				if err := s.enqueueAudit(ctx, tx, ports.AuditEvent{
					ID:        uuid.New(),
					UserID:    &sess.UserID,
					EventType: "refresh_reuse",
					IP:        ip,
					UserAgent: ua,
					Meta:      map[string]any{"family": sess.SessionFamilyID.String()},
					CreatedAt: now,
				}); err != nil {
					return err
				}
				metrics.AuthRefreshReuseTotal.Inc()
				return domain.ErrRefreshReuse
			}
			cmp := now
			if _, rerr := s.Store.RevokeSessionFamily(ctx, tx, sess.SessionFamilyID, "refresh_reuse_detected", now, &cmp); rerr != nil {
				return rerr
			}
			metrics.AuthRefreshReuseTotal.Inc()
			return domain.ErrRefreshReuse
		}
		if now.After(sess.ExpiresAt) {
			_ = s.Store.RevokeSession(ctx, tx, sess.ID, "expired", now)
			return domain.ErrInvalidRefresh
		}
		acc, err := s.Store.GetAccountByID(ctx, tx, sess.UserID)
		if err != nil {
			return err
		}
		if acc.Status != domain.StatusActive {
			return domain.ErrAccountBlocked
		}
		newRefresh, err := crypto.RandomRefreshToken()
		if err != nil {
			return err
		}
		newID := uuid.New()
		newHash := crypto.RefreshTokenHash(s.Pepper, newRefresh)
		exp := now.Add(s.RefreshTTL)
		newSess := ports.Session{
			ID:               newID,
			UserID:           sess.UserID,
			RefreshTokenHash: newHash,
			SessionFamilyID:  sess.SessionFamilyID,
			DeviceID:         mergeStr(sess.DeviceID, deviceID),
			DeviceName:       mergeStr(sess.DeviceName, nil),
			Platform:         sess.Platform,
			IP:               mergeStr(sess.IP, ip),
			UserAgent:        mergeStr(sess.UserAgent, ua),
			IssuedAt:         now,
			ExpiresAt:        exp,
			LastSeenAt:       ptrTime(now),
		}
		if err := s.Store.InsertSession(ctx, tx, newSess); err != nil {
			return err
		}
		if err := s.Store.RevokeSessionAsRotation(ctx, tx, sess.ID, newID, now); err != nil {
			return err
		}
		access, jti, aexp, err := s.JWT.SignAccess(ctx, sess.UserID, newID, acc.Role, s.AccessTTL)
		if err != nil {
			return err
		}
		_ = jti
		out = &ports.TokenPair{AccessToken: access, RefreshToken: newRefresh, AccessExp: aexp}
		if err := s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &sess.UserID,
			EventType: "session_rotated",
			IP:        ip,
			UserAgent: ua,
			Meta:      map[string]any{"old_session": sess.ID.String(), "new_session": newID.String()},
			CreatedAt: now,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := s.Redis.SetRefreshGrace(ctx, oldHash, *out, graceRefresh); err != nil {
		s.Log.Warn("refresh grace", "err", err)
	}
	metrics.AuthRefreshTotal.Inc()
	return out, nil
}

func (s *Service) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	cl, err := s.JWT.ParseAccess(ctx, accessToken)
	if err != nil {
		return domain.ErrBadRequest
	}
	ttl := time.Until(cl.Expires)
	if ttl > 0 {
		_ = s.Redis.Add(ctx, cl.JTI, ttl)
	}
	rh := crypto.RefreshTokenHash(s.Pepper, refreshToken)
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		sess, err := s.Store.GetSessionByRefreshHashForUpdate(ctx, tx, rh)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}
			return err
		}
		if sess.UserID != cl.Sub {
			return domain.ErrForbidden
		}
		now := time.Now().UTC()
		return s.Store.RevokeSession(ctx, tx, sess.ID, "logout", now)
	})
}

func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID, accessJTI string, accessExp time.Time) error {
	if accessJTI != "" {
		ttl := time.Until(accessExp)
		if ttl > 0 {
			_ = s.Redis.Add(ctx, accessJTI, ttl)
		}
	}
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		now := time.Now().UTC()
		_, err := s.Store.RevokeAllSessionsForUser(ctx, tx, userID, "logout_all", now)
		return err
	})
}

func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID) ([]ports.Session, error) {
	var list []ports.Session
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		list, err = s.Store.ListSessionsByUser(ctx, tx, userID)
		return err
	})
	return list, err
}

func (s *Service) DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		sess, err := s.Store.GetSessionByID(ctx, tx, sessionID)
		if err != nil {
			return err
		}
		if sess.UserID != userID {
			return domain.ErrForbidden
		}
		now := time.Now().UTC()
		return s.Store.RevokeSession(ctx, tx, sessionID, "logout", now)
	})
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	start := time.Now()
	_, norm, err := normalizeEmail(email)
	if err != nil {
		time.Sleep(forgotMinLatency)
		return nil
	}
	var acc *ports.Account
	_ = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		a, err := s.Store.GetAccountByEmailNorm(ctx, tx, norm)
		if err == nil {
			acc = a
		}
		return nil
	})
	verified := acc != nil && acc.EmailVerifiedAt != nil
	var emailTo string
	if acc != nil && acc.Email != nil {
		emailTo = *acc.Email
	}
	if acc != nil && acc.Status == domain.StatusActive && verified && emailTo != "" {
		salt, err := crypto.RandomSalt(16)
		if err != nil {
			s.Log.Warn("forgot password salt", "err", err)
		} else if plain, err := crypto.RandomDigits(6); err != nil {
			s.Log.Warn("forgot password code", "err", err)
		} else {
			hash := crypto.HashOTP(salt, plain)
			exp := time.Now().Add(s.ResetTTL)
			var resetID uuid.UUID
			if err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
				if err := s.Store.CancelPendingResets(ctx, tx, acc.ID); err != nil {
					return err
				}
				pr, err := s.Store.InsertPasswordReset(ctx, tx, acc.ID, emailTo, salt, hash, exp)
				if err != nil {
					return err
				}
				resetID = pr.ID
				return nil
			}); err != nil {
				s.Log.Warn("forgot password persist", "err", err)
			} else {
				s.Log.Info("password_reset_ready", "reset_id", resetID.String(), "email", emailTo)
				_ = s.Mail.SendPasswordResetOTP(ctx, emailTo, plain)
			}
		}
	}
	if d := time.Since(start); d < forgotMinLatency {
		time.Sleep(forgotMinLatency - d)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, resetID uuid.UUID, code, newPassword string) error {
	hash, err := s.Hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		r, err := s.Store.GetPasswordResetByIDForUpdate(ctx, tx, resetID)
		if err != nil {
			return err
		}
		if r.Status != "pending" {
			return domain.ErrInvalidOTP
		}
		if time.Now().After(r.ExpiresAt) {
			return domain.ErrInvalidOTP
		}
		if crypto.HashOTP(r.CodeSalt, code) != r.CodeHash {
			_ = s.Store.IncrementPasswordResetAttempt(ctx, tx, r.ID)
			return domain.ErrInvalidOTP
		}
		now := time.Now().UTC()
		if err := s.Store.MarkPasswordResetUsed(ctx, tx, r.ID, now); err != nil {
			return err
		}
		if err := s.Store.UpdatePasswordHash(ctx, tx, r.UserID, hash, 1, now); err != nil {
			return err
		}
		if _, err := s.Store.RevokeAllSessionsForUser(ctx, tx, r.UserID, "password_reset", now); err != nil {
			return err
		}
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &r.UserID,
			EventType: "password_reset_completed",
			Meta:      map[string]any{},
			CreatedAt: now,
		})
	})
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, current, newPassword string, currentSession uuid.UUID) error {
	hash, err := s.Hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		cred, err := s.Store.GetCredential(ctx, tx, userID)
		if err != nil {
			return err
		}
		ok, err := s.Hasher.Verify(cred.PasswordHash, current)
		if err != nil || !ok {
			return domain.ErrInvalidCredentials
		}
		now := time.Now().UTC()
		if err := s.Store.UpdatePasswordHash(ctx, tx, userID, hash, 1, now); err != nil {
			return err
		}
		sessions, err := s.Store.ListSessionsByUser(ctx, tx, userID)
		if err != nil {
			return err
		}
		for _, se := range sessions {
			if se.RevokedAt != nil {
				continue
			}
			if se.ID == currentSession {
				continue
			}
			_ = s.Store.RevokeSession(ctx, tx, se.ID, "password_changed", now)
		}
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &userID,
			EventType: "password_changed",
			Meta:      map[string]any{},
			CreatedAt: now,
		})
	})
}

func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		now := time.Now().UTC()
		if err := s.Store.SoftDeleteAccount(ctx, tx, userID, now); err != nil {
			return err
		}
		_, err := s.Store.RevokeAllSessionsForUser(ctx, tx, userID, "logout_all", now)
		return err
	})
}

func (s *Service) AdminBlock(ctx context.Context, actorID, target uuid.UUID, reason string) error {
	return s.adminWrite(ctx, actorID, target, func(ctx context.Context, tx pgx.Tx) error {
		return s.Store.SetBlocked(ctx, tx, target, reason, time.Now().UTC())
	}, "admin_block", map[string]any{"reason": reason})
}

func (s *Service) AdminUnblock(ctx context.Context, actorID, target uuid.UUID) error {
	return s.adminWrite(ctx, actorID, target, func(ctx context.Context, tx pgx.Tx) error {
		return s.Store.ClearBlocked(ctx, tx, target)
	}, "admin_unblock", map[string]any{})
}

func (s *Service) AdminForceLogout(ctx context.Context, actorID, target uuid.UUID) error {
	return s.adminWrite(ctx, actorID, target, func(ctx context.Context, tx pgx.Tx) error {
		_, err := s.Store.RevokeAllSessionsForUser(ctx, tx, target, "admin_force_logout", time.Now().UTC())
		return err
	}, "admin_force_logout", map[string]any{})
}

func (s *Service) adminWrite(ctx context.Context, actorID, target uuid.UUID, fn func(context.Context, pgx.Tx) error, event string, meta map[string]any) error {
	return s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		actor, err := s.Store.GetAccountByID(ctx, tx, actorID)
		if err != nil {
			return err
		}
		if actor.Role != domain.RoleAdmin {
			return domain.ErrForbidden
		}
		if err := fn(ctx, tx); err != nil {
			return err
		}
		return s.enqueueAudit(ctx, tx, ports.AuditEvent{
			ID:        uuid.New(),
			UserID:    &target,
			EventType: event,
			Meta:      mergeMeta(meta, "actor", actorID.String()),
			CreatedAt: time.Now().UTC(),
		})
	})
}

func mergeMeta(m map[string]any, k, v string) map[string]any {
	if m == nil {
		m = map[string]any{}
	}
	m[k] = v
	return m
}

func (s *Service) AdminListAudit(ctx context.Context, actorID, target uuid.UUID, limit, offset int) ([]ports.AuditEvent, error) {
	var list []ports.AuditEvent
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		actor, err := s.Store.GetAccountByID(ctx, tx, actorID)
		if err != nil {
			return err
		}
		if actor.Role != domain.RoleAdmin {
			return domain.ErrForbidden
		}
		list, err = s.Store.ListAuditByUser(ctx, tx, target, limit, offset)
		return err
	})
	return list, err
}
