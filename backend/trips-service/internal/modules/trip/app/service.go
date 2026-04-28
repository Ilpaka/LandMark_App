package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/domain"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/ports"
)

type Service struct{ Store ports.Store }

func New(s ports.Store) *Service { return &Service{Store: s} }

func (s *Service) CreateDraft(ctx context.Context, ownerID uuid.UUID, title string) (*domain.Trip, error) {
	if title == "" {
		return nil, domain.ErrBadRequest
	}
	t := domain.Trip{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Title:     title,
		Status:    domain.TripDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.Store.InsertTrip(ctx, t)
}

func (s *Service) List(ctx context.Context, ownerID uuid.UUID, status string, cursor string, limit int) ([]domain.Trip, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.Store.ListTrips(ctx, ownerID, status, cursor, limit)
}

func (s *Service) Get(ctx context.Context, ownerID, tripID uuid.UUID) (*domain.Trip, error) {
	t, err := s.Store.GetTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.ErrNotFound
	}
	if t.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	return t, nil
}

func (s *Service) Update(ctx context.Context, ownerID, tripID uuid.UUID, updates map[string]any) (*domain.Trip, error) {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return nil, err
	}
	return s.Store.UpdateTrip(ctx, tripID, updates)
}

func (s *Service) Delete(ctx context.Context, ownerID, tripID uuid.UUID) error {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return err
	}
	return s.Store.SetTripStatus(ctx, tripID, domain.TripArchived)
}

func (s *Service) AddStop(ctx context.Context, ownerID, tripID uuid.UUID, stop domain.TripStop) (*domain.TripStop, error) {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return nil, err
	}
	stop.ID = uuid.New()
	stop.TripID = tripID
	stop.CreatedAt = time.Now()
	return s.Store.InsertStop(ctx, stop)
}

func (s *Service) ListStops(ctx context.Context, ownerID, tripID uuid.UUID) ([]domain.TripStop, error) {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return nil, err
	}
	return s.Store.ListStops(ctx, tripID)
}

func (s *Service) MarkVisited(ctx context.Context, ownerID uuid.UUID, stopID uuid.UUID, at time.Time) error {
	return s.Store.MarkStopVisited(ctx, stopID, at)
}
