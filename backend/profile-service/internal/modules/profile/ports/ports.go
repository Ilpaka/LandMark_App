package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/domain"
)

type Store interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	UpsertProfile(ctx context.Context, p domain.Profile) (*domain.Profile, error)
	NicknameExists(ctx context.Context, normalized string, excludeUserID uuid.UUID) (bool, error)
	GetPrivacy(ctx context.Context, userID uuid.UUID) (*domain.PrivacySettings, error)
	UpsertPrivacy(ctx context.Context, s domain.PrivacySettings) (*domain.PrivacySettings, error)

	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)
	ListFollowers(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Profile, error)
	ListFollowing(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Profile, error)
	FollowCounts(ctx context.Context, userID uuid.UUID) (followers int, following int, err error)
}
