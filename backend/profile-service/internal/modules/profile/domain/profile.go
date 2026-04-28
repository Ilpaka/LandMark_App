package domain

import (
	"time"
	"github.com/google/uuid"
)

type Profile struct {
	UserID      uuid.UUID  `json:"user_id"`
	Nickname    string     `json:"nickname"`
	DisplayName string     `json:"display_name"`
	Bio         *string    `json:"bio,omitempty"`
	AvatarMediaID *uuid.UUID `json:"avatar_media_id,omitempty"`
	City        *string    `json:"city,omitempty"`
	Country     *string    `json:"country,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PrivacySettings struct {
	UserID             uuid.UUID `json:"user_id"`
	ProfileVisibility  string    `json:"profile_visibility"`
	TripsVisibility    string    `json:"trips_visibility"`
	JournalVisibility  string    `json:"journal_visibility"`
	AnalyticsEnabled   bool      `json:"analytics_enabled"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type UpdateProfileInput struct {
	Nickname    *string
	DisplayName *string
	Bio         *string
	City        *string
	Country     *string
}

type UpdatePrivacyInput struct {
	ProfileVisibility *string
	TripsVisibility   *string
	JournalVisibility *string
	AnalyticsEnabled  *bool
}
