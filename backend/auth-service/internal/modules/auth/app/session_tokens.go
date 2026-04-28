package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// sessionDeviceInfo carries optional client metadata for a new session.
type sessionDeviceInfo struct {
	DeviceID   *string
	DeviceName *string
	Platform   *string
	IP         *string
	UserAgent  *string
}

func (s *Service) issueFreshSession(ctx context.Context, tx pgx.Tx, acc *ports.Account, dev sessionDeviceInfo, now time.Time) (*ports.TokenPair, uuid.UUID, error) {
	if err := s.Store.UpdateLastLogin(ctx, tx, acc.ID, now); err != nil {
		return nil, uuid.Nil, err
	}
	refresh, err := crypto.RandomRefreshToken()
	if err != nil {
		return nil, uuid.Nil, err
	}
	sid := uuid.New()
	fam := uuid.New()
	rh := crypto.RefreshTokenHash(s.Pepper, refresh)
	sess := ports.Session{
		ID:               sid,
		UserID:           acc.ID,
		RefreshTokenHash: rh,
		SessionFamilyID:  fam,
		DeviceID:         dev.DeviceID,
		DeviceName:       dev.DeviceName,
		Platform:         dev.Platform,
		IP:               dev.IP,
		UserAgent:        dev.UserAgent,
		IssuedAt:         now,
		ExpiresAt:        now.Add(s.RefreshTTL),
		LastSeenAt:       ptrTime(now),
	}
	if err := s.Store.InsertSession(ctx, tx, sess); err != nil {
		return nil, uuid.Nil, err
	}
	access, _, aexp, err := s.JWT.SignAccess(ctx, acc.ID, sid, acc.Role, s.AccessTTL)
	if err != nil {
		return nil, uuid.Nil, err
	}
	return &ports.TokenPair{AccessToken: access, RefreshToken: refresh, AccessExp: aexp}, sid, nil
}
