package domain

import (
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID         uuid.UUID   `json:"id"`
	AuthorID   uuid.UUID   `json:"author_id"`
	TripID     *uuid.UUID  `json:"trip_id"`
	PlaceID    *uuid.UUID  `json:"place_id"`
	Latitude   *float64    `json:"latitude"`
	Longitude  *float64    `json:"longitude"`
	Title      *string     `json:"title"`
	Body       string      `json:"body"`
	Mood       *string     `json:"mood"`
	Rating     *int        `json:"rating"`
	OccurredAt time.Time   `json:"occurred_at"`
	MediaIDs   []uuid.UUID `json:"media_ids"`
	Tags       []string    `json:"tags"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	DeletedAt  *time.Time  `json:"deleted_at,omitempty"`
}

type Reaction struct {
	EntryID   uuid.UUID `json:"entry_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type ReactionCount struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}
