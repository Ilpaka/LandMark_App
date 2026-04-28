package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/domain"
)

type Store interface {
	InsertNotification(ctx context.Context, in domain.SendInput) (*domain.Notification, error)
	ListNotifications(ctx context.Context, f domain.ListFilter) (*domain.NotificationsPage, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	UnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	UpsertDevice(ctx context.Context, d domain.Device) (*domain.Device, error)
	DeleteDevice(ctx context.Context, id, userID uuid.UUID) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.Preferences, error)
	UpsertPreferences(ctx context.Context, p domain.Preferences) (*domain.Preferences, error)
}
