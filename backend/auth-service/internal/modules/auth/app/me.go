package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

// MeInfo is the extended /me payload (DB-backed fields).
type MeInfo struct {
	UserID          uuid.UUID
	HasPassword     bool
	Email           *string
	EmailVerified   bool
	PhoneE164       *string
	PhoneVerified   bool
}

// Me loads account flags for the authenticated user.
func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*MeInfo, error) {
	var out *MeInfo
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.Store.GetAccountByID(ctx, tx, userID)
		if err != nil {
			return err
		}
		if acc.Status == domain.StatusDeleted {
			return domain.ErrAccountDeleted
		}
		if acc.Status == domain.StatusBlocked {
			return domain.ErrAccountBlocked
		}
		_, cerr := s.Store.GetCredential(ctx, tx, userID)
		hasPwd := cerr == nil
		if cerr != nil && !errors.Is(cerr, domain.ErrNotFound) {
			return cerr
		}
		ev := acc.EmailVerifiedAt != nil
		pv := acc.PhoneVerifiedAt != nil
		out = &MeInfo{
			UserID:        acc.ID,
			HasPassword:   hasPwd,
			Email:         acc.Email,
			EmailVerified: ev,
			PhoneE164:     acc.PhoneE164,
			PhoneVerified: pv,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
