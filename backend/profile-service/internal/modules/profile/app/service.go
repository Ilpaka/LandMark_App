package app

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/domain"
	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/ports"
)

type Service struct {
	Store ports.Store
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	return s.Store.GetProfile(ctx, userID)
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, in domain.UpdateProfileInput) (*domain.Profile, error) {
	p, err := s.Store.GetProfile(ctx, userID)
	if err == domain.ErrNotFound {
		p = &domain.Profile{UserID: userID, Nickname: userID.String()[:8], DisplayName: "User"}
	} else if err != nil {
		return nil, err
	}

	if in.Nickname != nil {
		nick := strings.TrimSpace(*in.Nickname)
		if nick == "" {
			return nil, domain.ErrBadRequest
		}
		norm := strings.ToLower(nick)
		exists, err := s.Store.NicknameExists(ctx, norm, userID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrConflict
		}
		p.Nickname = nick
	}
	if in.DisplayName != nil {
		if strings.TrimSpace(*in.DisplayName) == "" {
			return nil, domain.ErrBadRequest
		}
		p.DisplayName = strings.TrimSpace(*in.DisplayName)
	}
	if in.Bio != nil {
		p.Bio = in.Bio
	}
	if in.City != nil {
		p.City = in.City
	}
	if in.Country != nil {
		p.Country = in.Country
	}
	return s.Store.UpsertProfile(ctx, *p)
}

func (s *Service) GetPrivacy(ctx context.Context, userID uuid.UUID) (*domain.PrivacySettings, error) {
	priv, err := s.Store.GetPrivacy(ctx, userID)
	if err == domain.ErrNotFound {
		return &domain.PrivacySettings{
			UserID: userID, ProfileVisibility: "public",
			TripsVisibility: "private", JournalVisibility: "private",
			AnalyticsEnabled: true,
		}, nil
	}
	return priv, err
}

func (s *Service) UpdatePrivacy(ctx context.Context, userID uuid.UUID, in domain.UpdatePrivacyInput) (*domain.PrivacySettings, error) {
	priv, err := s.GetPrivacy(ctx, userID)
	if err != nil {
		return nil, err
	}
	if in.ProfileVisibility != nil {
		priv.ProfileVisibility = *in.ProfileVisibility
	}
	if in.TripsVisibility != nil {
		priv.TripsVisibility = *in.TripsVisibility
	}
	if in.JournalVisibility != nil {
		priv.JournalVisibility = *in.JournalVisibility
	}
	if in.AnalyticsEnabled != nil {
		priv.AnalyticsEnabled = *in.AnalyticsEnabled
	}
	return s.Store.UpsertPrivacy(ctx, *priv)
}
