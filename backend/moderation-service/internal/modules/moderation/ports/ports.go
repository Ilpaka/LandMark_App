package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/domain"
)

type Store interface {
	InsertItem(ctx context.Context, targetType domain.TargetType, targetID, submittedBy uuid.UUID) (*domain.QueueItem, error)
	GetItem(ctx context.Context, id uuid.UUID) (*domain.QueueItem, error)
	GetItemByTarget(ctx context.Context, targetType domain.TargetType, targetID uuid.UUID) (*domain.QueueItem, error)
	ListPending(ctx context.Context, f domain.ListFilter) (*domain.QueuePage, error)
	UpdateDecision(ctx context.Context, id uuid.UUID, status domain.ItemStatus, moderatorID uuid.UUID, note *string) error
}

type PlacesClient interface {
	ApprovePlace(ctx context.Context, placeID uuid.UUID) error
	RejectPlace(ctx context.Context, placeID uuid.UUID, reason string) error
}
