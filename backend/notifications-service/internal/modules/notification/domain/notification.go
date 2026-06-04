package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	DeepLink  *string         `json:"deep_link,omitempty"`
	Meta      json.RawMessage `json:"meta"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type Device struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Platform   string    `json:"platform"`
	PushToken  string    `json:"push_token"`
	LastSeenAt time.Time `json:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type Preferences struct {
	UserID                uuid.UUID `json:"user_id"`
	PushTripReminders     bool      `json:"push_trip_reminders"`
	PushModerationResult  bool      `json:"push_moderation_result"`
	PushSyncStatus        bool      `json:"push_sync_status"`
	PushMarketing         bool      `json:"push_marketing"`
	EmailWelcome          bool      `json:"email_welcome"`
	EmailModerationResult bool      `json:"email_moderation_result"`
	EmailMarketing        bool      `json:"email_marketing"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type ListFilter struct {
	UserID     uuid.UUID
	UnreadOnly bool
	Cursor     string
	Limit      int
}

type NotificationsPage struct {
	Items      []Notification `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

type SendInput struct {
	UserID   uuid.UUID
	Type     string
	Title    string
	Body     string
	DeepLink *string
	Meta     map[string]any
}
