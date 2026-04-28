package domain

import (
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID         uuid.UUID
	AuthorID   uuid.UUID
	TripID     *uuid.UUID
	PlaceID    *uuid.UUID
	Latitude   *float64
	Longitude  *float64
	Title      *string
	Body       string
	Mood       *string
	Rating     *int
	OccurredAt time.Time
	MediaIDs   []uuid.UUID
	Tags       []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
