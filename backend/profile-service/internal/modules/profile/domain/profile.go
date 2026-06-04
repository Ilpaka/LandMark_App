package domain

import (
	"github.com/google/uuid"
	"time"
)

type Profile struct {
	UserID        uuid.UUID  `json:"user_id"`
	Nickname      string     `json:"nickname"`
	DisplayName   string     `json:"display_name"`
	Bio           *string    `json:"bio,omitempty"`
	AvatarMediaID *uuid.UUID `json:"avatar_media_id,omitempty"`
	City          *string    `json:"city,omitempty"`
	Country       *string    `json:"country,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type PrivacySettings struct {
	UserID            uuid.UUID `json:"user_id"`
	ProfileVisibility string    `json:"profile_visibility"`
	TripsVisibility   string    `json:"trips_visibility"`
	JournalVisibility string    `json:"journal_visibility"`
	AnalyticsEnabled  bool      `json:"analytics_enabled"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UpdateProfileInput struct {
	Nickname      *string
	DisplayName   *string
	Bio           *string
	City          *string
	Country       *string
	AvatarMediaID *uuid.UUID
}

type UpdatePrivacyInput struct {
	ProfileVisibility *string
	TripsVisibility   *string
	JournalVisibility *string
	AnalyticsEnabled  *bool
}

type Follow struct {
	FollowerID uuid.UUID `json:"follower_id"`
	FolloweeID uuid.UUID `json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowStats struct {
	FollowersCount int  `json:"followers_count"`
	FollowingCount int  `json:"following_count"`
	IsFollowing    bool `json:"is_following"`
}
