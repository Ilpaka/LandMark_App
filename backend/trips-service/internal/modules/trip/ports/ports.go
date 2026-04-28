package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/domain"
)

type Store interface {
	InsertTrip(ctx context.Context, t domain.Trip) (*domain.Trip, error)
	ListTrips(ctx context.Context, ownerID uuid.UUID, status, cursor string, limit int) ([]domain.Trip, string, error)
	GetTrip(ctx context.Context, id uuid.UUID) (*domain.Trip, error)
	UpdateTrip(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.Trip, error)
	SetTripStatus(ctx context.Context, id uuid.UUID, status domain.TripStatus) error
	InsertStop(ctx context.Context, s domain.TripStop) (*domain.TripStop, error)
	ListStops(ctx context.Context, tripID uuid.UUID) ([]domain.TripStop, error)
	UpdateStop(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.TripStop, error)
	DeleteStop(ctx context.Context, id uuid.UUID) error
	MarkStopVisited(ctx context.Context, id uuid.UUID, at time.Time) error
}
