package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/domain"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/ports"
)

type Service struct {
	Store ports.Store
}

func New(store ports.Store) *Service { return &Service{Store: store} }

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.Store.ListCategories(ctx)
}

func (s *Service) ListPlaces(ctx context.Context, f domain.ListFilter) ([]domain.Place, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 200
	}
	return s.Store.ListPlaces(ctx, f)
}

func (s *Service) GetPlace(ctx context.Context, id uuid.UUID, userID *uuid.UUID, role string) (*domain.Place, error) {
	p, err := s.Store.GetPlace(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}
	// Guest and regular user only see published (or their own)
	if p.Status != domain.StatusPublished {
		if userID == nil {
			return nil, domain.ErrNotFound
		}
		isOwner := p.AuthorID != nil && *p.AuthorID == *userID
		if !isOwner && role != "admin" {
			return nil, domain.ErrNotFound
		}
	}
	return p, nil
}

func (s *Service) CreateDraft(ctx context.Context, userID uuid.UUID, title string, lat, lng float64) (*domain.Place, error) {
	if title == "" {
		return nil, domain.ErrBadRequest
	}
	p := domain.Place{
		ID:        uuid.New(),
		Title:     title,
		Latitude:  lat,
		Longitude: lng,
		AuthorID:  &userID,
		Status:    domain.StatusDraft,
		Source:    "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.Store.InsertPlace(ctx, p)
}

func (s *Service) UpdateDraft(ctx context.Context, userID, placeID uuid.UUID, updates map[string]any) (*domain.Place, error) {
	p, err := s.Store.GetPlace(ctx, placeID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}
	if p.AuthorID == nil || *p.AuthorID != userID {
		return nil, domain.ErrForbidden
	}
	if p.Status != domain.StatusDraft {
		return nil, domain.ErrConflict
	}
	return s.Store.UpdatePlace(ctx, placeID, updates)
}

func (s *Service) Submit(ctx context.Context, userID, placeID uuid.UUID) error {
	p, err := s.Store.GetPlace(ctx, placeID)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrNotFound
	}
	if p.AuthorID == nil || *p.AuthorID != userID {
		return domain.ErrForbidden
	}
	if p.Status != domain.StatusDraft {
		return domain.ErrConflict
	}
	if err := s.Store.SetPlaceStatus(ctx, placeID, domain.StatusPendingModeration, nil); err != nil {
		return err
	}
	return s.Store.InsertOutbox(ctx, "place", "place_submitted", map[string]any{
		"place_id":     placeID.String(),
		"author_id":    userID.String(),
		"submitted_at": time.Now(),
	})
}

func (s *Service) ApprovePlace(ctx context.Context, placeID, adminID uuid.UUID) error {
	now := time.Now()
	if err := s.Store.SetPlaceStatus(ctx, placeID, domain.StatusPublished, map[string]any{"published_at": now}); err != nil {
		return err
	}
	p, _ := s.Store.GetPlace(ctx, placeID)
	if p != nil && p.AuthorID != nil {
		_ = s.Store.InsertOutbox(ctx, "place", "place_published", map[string]any{
			"place_id":     placeID.String(),
			"author_id":    p.AuthorID.String(),
			"published_at": now,
		})
	}
	return nil
}

func (s *Service) RejectPlace(ctx context.Context, placeID, adminID uuid.UUID, reason string) error {
	if err := s.Store.SetPlaceStatus(ctx, placeID, domain.StatusRejected, map[string]any{"reject_reason": reason}); err != nil {
		return err
	}
	p, _ := s.Store.GetPlace(ctx, placeID)
	if p != nil && p.AuthorID != nil {
		_ = s.Store.InsertOutbox(ctx, "place", "place_rejected", map[string]any{
			"place_id":  placeID.String(),
			"author_id": p.AuthorID.String(),
			"reason":    reason,
		})
	}
	return nil
}

func (s *Service) CreateCategory(ctx context.Context, slug, title, icon, color string, sortOrder int) (*domain.Category, error) {
	c := domain.Category{
		ID:        uuid.New(),
		Slug:      slug,
		Title:     title,
		Icon:      icon,
		Color:     color,
		SortOrder: sortOrder,
		Active:    true,
		CreatedAt: time.Now(),
	}
	return s.Store.InsertCategory(ctx, c)
}
