package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/media-service/internal/modules/media/domain"
)

type Store interface {
	InsertUpload(ctx context.Context, u domain.Upload) (*domain.Upload, error)
	GetUpload(ctx context.Context, id uuid.UUID) (*domain.Upload, error)
	FinalizeUpload(ctx context.Context, id uuid.UUID) error
	InsertMedia(ctx context.Context, m domain.Media) (*domain.Media, error)
	GetMedia(ctx context.Context, id uuid.UUID) (*domain.Media, error)
	SoftDeleteMedia(ctx context.Context, id, ownerID uuid.UUID) error
}

type ObjectStore interface {
	PresignedPutURL(ctx context.Context, objectKey, mime string, ttl time.Duration) (string, error)
	PresignedGetURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error)
	PublicURL(objectKey string) string
}
