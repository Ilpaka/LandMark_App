package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/favorites-service/internal/modules/favorite/domain"
)

type Store interface {
	// Collections
	GetDefaultCollection(ctx context.Context, userID uuid.UUID) (*domain.Collection, error)
	GetOrCreateDefaultCollection(ctx context.Context, userID uuid.UUID) (*domain.Collection, error)
	ListCollections(ctx context.Context, userID uuid.UUID) ([]domain.Collection, error)
	GetCollection(ctx context.Context, id uuid.UUID) (*domain.Collection, error)
	InsertCollection(ctx context.Context, userID uuid.UUID, title, color string) (*domain.Collection, error)
	UpdateCollection(ctx context.Context, id uuid.UUID, title, color string) (*domain.Collection, error)
	DeleteCollection(ctx context.Context, id uuid.UUID) error

	// Favorites
	ListFavorites(ctx context.Context, f domain.ListFavoritesFilter) (*domain.FavoritesPage, error)
	GetFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID) (*domain.Favorite, error)
	UpsertFavorite(ctx context.Context, fav domain.Favorite) error
	DeleteFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID) error
}
