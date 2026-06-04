package domain

import (
	"github.com/google/uuid"
	"time"
)

type TargetType string

const (
	TargetPlace TargetType = "place"
	TargetTrip  TargetType = "trip"
	TargetEntry TargetType = "entry"
)

type Collection struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Color     string    `json:"color"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type Favorite struct {
	UserID       uuid.UUID  `json:"user_id"`
	TargetType   TargetType `json:"target_type"`
	TargetID     uuid.UUID  `json:"target_id"`
	CollectionID *uuid.UUID `json:"collection_id,omitempty"`
	Note         *string    `json:"note,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ListFavoritesFilter struct {
	UserID       uuid.UUID
	TargetType   *TargetType
	CollectionID *uuid.UUID
	Cursor       string
	Limit        int
}

type FavoritesPage struct {
	Items      []Favorite `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
}
