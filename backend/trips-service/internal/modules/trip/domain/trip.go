package domain

import (
	"time"

	"github.com/google/uuid"
)

type TripStatus string

const (
	TripDraft      TripStatus = "draft"
	TripPlanned    TripStatus = "planned"
	TripInProgress TripStatus = "in_progress"
	TripCompleted  TripStatus = "completed"
	TripArchived   TripStatus = "archived"
)

type Trip struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	Title        string
	Subtitle     *string
	StartDate    *time.Time
	EndDate      *time.Time
	CoverMediaID *uuid.UUID
	Status       TripStatus
	Region       *string
	CenterLat    *float64
	CenterLng    *float64
	StopsCount   int
	EntriesCount int
	PhotosCount  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TripStop struct {
	ID        uuid.UUID
	TripID    uuid.UUID
	PlaceID   *uuid.UUID
	Title     string
	Latitude  float64
	Longitude float64
	PlannedAt *time.Time
	VisitedAt *time.Time
	Notes     *string
	SortOrder int
	CreatedAt time.Time
}
