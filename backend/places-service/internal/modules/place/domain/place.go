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
	ID        uuid.UUID
	Slug      string
	Title     string
	Icon      string
	Color     string
	SortOrder int
	Active    bool
	CreatedAt time.Time
}

type Place struct {
	ID           uuid.UUID
	Title        string
	Description  string
	Latitude     float64
	Longitude    float64
	Address      *string
	City         *string
	Country      *string
	AuthorID     *uuid.UUID
	Status       PlaceStatus
	RejectReason *string
	Source       string
	CoverMediaID *uuid.UUID
	CategoryIDs  []uuid.UUID
	PublishedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
