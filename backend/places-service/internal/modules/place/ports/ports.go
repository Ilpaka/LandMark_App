package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/domain"
)

type Store interface {
	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error)
	InsertCategory(ctx context.Context, c domain.Category) (*domain.Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, title, icon, color string, sortOrder int) (*domain.Category, error)
	DeactivateCategory(ctx context.Context, id uuid.UUID) error

	ListPlaces(ctx context.Context, f domain.ListFilter) ([]domain.Place, error)
	GetPlace(ctx context.Context, id uuid.UUID) (*domain.Place, error)
	InsertPlace(ctx context.Context, p domain.Place) (*domain.Place, error)
	UpdatePlace(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.Place, error)
	SetPlaceStatus(ctx context.Context, id uuid.UUID, status domain.PlaceStatus, extra map[string]any) error
	SetPlaceCategories(ctx context.Context, placeID uuid.UUID, categoryIDs []uuid.UUID) error
	InsertOutbox(ctx context.Context, aggType, evType string, payload map[string]any) error
}
