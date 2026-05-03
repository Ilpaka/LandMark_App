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
	ID           uuid.UUID  `json:"id"`
	OwnerID      uuid.UUID  `json:"owner_id"`
	Title        string     `json:"title"`
	Subtitle     *string    `json:"subtitle"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	CoverMediaID *uuid.UUID `json:"cover_media_id"`
	Status       TripStatus `json:"status"`
	Region       *string    `json:"region"`
	CenterLat    *float64   `json:"center_lat"`
	CenterLng    *float64   `json:"center_lng"`
	StopsCount   int        `json:"stops_count"`
	EntriesCount int        `json:"entries_count"`
	PhotosCount  int        `json:"photos_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type TripStop struct {
	ID        uuid.UUID  `json:"id"`
	TripID    uuid.UUID  `json:"trip_id"`
	PlaceID   *uuid.UUID `json:"place_id"`
	Title     string     `json:"title"`
	Latitude  float64    `json:"latitude"`
	Longitude float64    `json:"longitude"`
	PlannedAt *time.Time `json:"planned_at"`
	VisitedAt *time.Time `json:"visited_at"`
	Notes     *string    `json:"notes"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
}
