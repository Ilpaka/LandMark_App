package app

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/favorites-service/internal/modules/favorite/domain"
	"github.com/ilpaka/landmark_app/backend/favorites-service/internal/modules/favorite/ports"
)

type Service struct {
	Store ports.Store
}

func (s *Service) ListCollections(ctx context.Context, userID uuid.UUID) ([]domain.Collection, error) {
	return s.Store.ListCollections(ctx, userID)
}

func (s *Service) CreateCollection(ctx context.Context, userID uuid.UUID, title, color string) (*domain.Collection, error) {
	if title == "" {
		return nil, domain.ErrBadRequest
	}
	if color == "" {
		color = "#0E7C7B"
	}
	return s.Store.InsertCollection(ctx, userID, title, color)
}

func (s *Service) UpdateCollection(ctx context.Context, userID, id uuid.UUID, title, color string) (*domain.Collection, error) {
	col, err := s.Store.GetCollection(ctx, id)
	if err != nil {
		return nil, err
	}
	if col.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if title != "" {
		col.Title = title
	}
	if color != "" {
		col.Color = color
	}
	return s.Store.UpdateCollection(ctx, id, col.Title, col.Color)
}

func (s *Service) DeleteCollection(ctx context.Context, userID, id uuid.UUID) error {
	col, err := s.Store.GetCollection(ctx, id)
	if err != nil {
		return err
	}
	if col.UserID != userID {
		return domain.ErrForbidden
	}
	if col.IsDefault {
		return domain.ErrBadRequest
	}
	return s.Store.DeleteCollection(ctx, id)
}

func (s *Service) ListFavorites(ctx context.Context, f domain.ListFavoritesFilter) (*domain.FavoritesPage, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.Store.ListFavorites(ctx, f)
}

func (s *Service) AddFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID, collectionID *uuid.UUID, note *string) error {
	switch tt {
	case domain.TargetPlace, domain.TargetTrip, domain.TargetEntry:
	default:
		return domain.ErrBadRequest
	}
	if collectionID == nil {
		col, err := s.Store.GetOrCreateDefaultCollection(ctx, userID)
		if err != nil {
			return err
		}
		collectionID = &col.ID
	} else {
		col, err := s.Store.GetCollection(ctx, *collectionID)
		if err != nil {
			return err
		}
		if col.UserID != userID {
			return domain.ErrForbidden
		}
	}
	return s.Store.UpsertFavorite(ctx, domain.Favorite{
		UserID: userID, TargetType: tt, TargetID: targetID,
		CollectionID: collectionID, Note: note,
	})
}

func (s *Service) RemoveFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID) error {
	return s.Store.DeleteFavorite(ctx, userID, tt, targetID)
}
