package domain

import (
	"time"

	"github.com/google/uuid"
)

type PlaceStatus string

const (
	StatusDraft             PlaceStatus = "draft"
	StatusPendingModeration PlaceStatus = "pending_moderation"
	StatusPublished         PlaceStatus = "published"
	StatusRejected          PlaceStatus = "rejected"
	StatusArchived          PlaceStatus = "archived"
)

type Category struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Icon      string    `json:"icon"`
	Color     string    `json:"color"`
	SortOrder int       `json:"sort_order"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type Place struct {
	ID           uuid.UUID   `json:"id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	Latitude     float64     `json:"latitude"`
	Longitude    float64     `json:"longitude"`
	Address      *string     `json:"address"`
	City         *string     `json:"city"`
	Country      *string     `json:"country"`
	AuthorID     *uuid.UUID  `json:"author_id"`
	Status       PlaceStatus `json:"status"`
	RejectReason *string     `json:"reject_reason"`
	Source       string      `json:"source"`
	CoverMediaID *uuid.UUID  `json:"cover_media_id"`
	CategoryIDs  []uuid.UUID `json:"category_ids"`
	PublishedAt  *time.Time  `json:"published_at"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type BBox struct {
	SouthLat float64
	WestLng  float64
	NorthLat float64
	EastLng  float64
}

type ListFilter struct {
	BBox         *BBox
	CategorySlug string
	Q            string
	Limit        int
}
