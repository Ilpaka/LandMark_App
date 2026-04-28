package app

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/domain"
	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/ports"
)

type Service struct {
	Store ports.Store
}

func (s *Service) Send(ctx context.Context, in domain.SendInput) (*domain.Notification, error) {
	if in.Title == "" || in.Body == "" {
		return nil, domain.ErrBadRequest
	}
	return s.Store.InsertNotification(ctx, in)
}

func (s *Service) List(ctx context.Context, f domain.ListFilter) (*domain.NotificationsPage, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.Store.ListNotifications(ctx, f)
}

func (s *Service) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.Store.MarkRead(ctx, id, userID)
}

func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.Store.MarkAllRead(ctx, userID)
}

func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.Store.UnreadCount(ctx, userID)
}

func (s *Service) UpsertDevice(ctx context.Context, userID uuid.UUID, platform, token string) (*domain.Device, error) {
	switch platform {
	case "ios", "android":
	default:
		return nil, domain.ErrBadRequest
	}
	return s.Store.UpsertDevice(ctx, domain.Device{
		ID: uuid.New(), UserID: userID, Platform: platform, PushToken: token,
	})
}

func (s *Service) DeleteDevice(ctx context.Context, id, userID uuid.UUID) error {
	return s.Store.DeleteDevice(ctx, id, userID)
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.Preferences, error) {
	prefs, err := s.Store.GetPreferences(ctx, userID)
	if err == domain.ErrNotFound {
		return &domain.Preferences{
			UserID: userID, PushTripReminders: true,
			PushModerationResult: true, EmailWelcome: true,
			EmailModerationResult: true,
		}, nil
	}
	return prefs, err
}

func (s *Service) UpdatePreferences(ctx context.Context, userID uuid.UUID, patch map[string]any) (*domain.Preferences, error) {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}
	if v, ok := patch["push_trip_reminders"]; ok { prefs.PushTripReminders = v.(bool) }
	if v, ok := patch["push_moderation_result"]; ok { prefs.PushModerationResult = v.(bool) }
	if v, ok := patch["push_sync_status"]; ok { prefs.PushSyncStatus = v.(bool) }
	if v, ok := patch["push_marketing"]; ok { prefs.PushMarketing = v.(bool) }
	if v, ok := patch["email_welcome"]; ok { prefs.EmailWelcome = v.(bool) }
	if v, ok := patch["email_moderation_result"]; ok { prefs.EmailModerationResult = v.(bool) }
	if v, ok := patch["email_marketing"]; ok { prefs.EmailMarketing = v.(bool) }
	return s.Store.UpsertPreferences(ctx, *prefs)
}
