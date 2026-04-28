package domain

import (
	"time"
	"github.com/google/uuid"
)

type ItemStatus string
const (
	StatusPending  ItemStatus = "pending"
	StatusApproved ItemStatus = "approved"
	StatusRejected ItemStatus = "rejected"
)

type TargetType string
const TargetPlace TargetType = "place"

type QueueItem struct {
	ID           uuid.UUID  `json:"id"`
	TargetType   TargetType `json:"target_type"`
	TargetID     uuid.UUID  `json:"target_id"`
	SubmittedBy  uuid.UUID  `json:"submitted_by"`
	Status       ItemStatus `json:"status"`
	ModeratorID  *uuid.UUID `json:"moderator_id,omitempty"`
	Note         *string    `json:"note,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ListFilter struct {
	Status *ItemStatus
	Cursor string
	Limit  int
}

type QueuePage struct {
	Items      []QueueItem `json:"items"`
	NextCursor string      `json:"next_cursor,omitempty"`
}

type ModerationStats struct {
	Pending  int `json:"pending"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
	Total    int `json:"total"`
}
