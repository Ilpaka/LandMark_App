package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/domain"
)

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT user_id,nickname,display_name,bio,avatar_media_id,city,country,created_at,updated_at
         FROM profile.profile_profiles WHERE user_id=$1`, userID)
	return scanProfile(row)
}

func (s *Store) UpsertProfile(ctx context.Context, p domain.Profile) (*domain.Profile, error) {
	norm := strings.ToLower(p.Nickname)
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO profile.profile_profiles(user_id,nickname,nickname_normalized,display_name,bio,city,country)
         VALUES($1,$2,$3,$4,$5,$6,$7)
         ON CONFLICT(user_id) DO UPDATE SET
           nickname=EXCLUDED.nickname,
           nickname_normalized=EXCLUDED.nickname_normalized,
           display_name=EXCLUDED.display_name,
           bio=EXCLUDED.bio,
           city=EXCLUDED.city,
           country=EXCLUDED.country,
           updated_at=now()
         RETURNING user_id,nickname,display_name,bio,avatar_media_id,city,country,created_at,updated_at`,
		p.UserID, p.Nickname, norm, p.DisplayName, p.Bio, p.City, p.Country)
	return scanProfile(row)
}

func (s *Store) NicknameExists(ctx context.Context, normalized string, excludeUserID uuid.UUID) (bool, error) {
	var exists bool
	err := s.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM profile.profile_profiles WHERE nickname_normalized=$1 AND user_id<>$2)`,
		normalized, excludeUserID).Scan(&exists)
	return exists, err
}

func (s *Store) GetPrivacy(ctx context.Context, userID uuid.UUID) (*domain.PrivacySettings, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT user_id,profile_visibility,trips_visibility,journal_visibility,analytics_enabled,updated_at
         FROM profile.profile_privacy_settings WHERE user_id=$1`, userID)
	var priv domain.PrivacySettings
	if err := row.Scan(&priv.UserID, &priv.ProfileVisibility, &priv.TripsVisibility,
		&priv.JournalVisibility, &priv.AnalyticsEnabled, &priv.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &priv, nil
}

func (s *Store) UpsertPrivacy(ctx context.Context, priv domain.PrivacySettings) (*domain.PrivacySettings, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO profile.profile_privacy_settings(user_id,profile_visibility,trips_visibility,journal_visibility,analytics_enabled)
         VALUES($1,$2,$3,$4,$5)
         ON CONFLICT(user_id) DO UPDATE SET
           profile_visibility=EXCLUDED.profile_visibility,
           trips_visibility=EXCLUDED.trips_visibility,
           journal_visibility=EXCLUDED.journal_visibility,
           analytics_enabled=EXCLUDED.analytics_enabled,
           updated_at=now()
         RETURNING user_id,profile_visibility,trips_visibility,journal_visibility,analytics_enabled,updated_at`,
		priv.UserID, priv.ProfileVisibility, priv.TripsVisibility,
		priv.JournalVisibility, priv.AnalyticsEnabled)
	var p domain.PrivacySettings
	if err := row.Scan(&p.UserID, &p.ProfileVisibility, &p.TripsVisibility,
		&p.JournalVisibility, &p.AnalyticsEnabled, &p.UpdatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO profile.profile_follows(follower_id,followee_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		followerID, followeeID)
	return err
}

func (s *Store) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	_, err := s.Pool.Exec(ctx,
		`DELETE FROM profile.profile_follows WHERE follower_id=$1 AND followee_id=$2`,
		followerID, followeeID)
	return err
}

func (s *Store) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	var exists bool
	err := s.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM profile.profile_follows WHERE follower_id=$1 AND followee_id=$2)`,
		followerID, followeeID).Scan(&exists)
	return exists, err
}

func (s *Store) ListFollowers(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Profile, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT p.user_id,p.nickname,p.display_name,p.bio,p.avatar_media_id,p.city,p.country,p.created_at,p.updated_at
         FROM profile.profile_follows f
         JOIN profile.profile_profiles p ON p.user_id=f.follower_id
         WHERE f.followee_id=$1 ORDER BY f.created_at DESC LIMIT $2`, userID, limit)
	return collectProfiles(rows, err)
}

func (s *Store) ListFollowing(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Profile, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT p.user_id,p.nickname,p.display_name,p.bio,p.avatar_media_id,p.city,p.country,p.created_at,p.updated_at
         FROM profile.profile_follows f
         JOIN profile.profile_profiles p ON p.user_id=f.followee_id
         WHERE f.follower_id=$1 ORDER BY f.created_at DESC LIMIT $2`, userID, limit)
	return collectProfiles(rows, err)
}

func (s *Store) FollowCounts(ctx context.Context, userID uuid.UUID) (followers int, following int, err error) {
	err = s.Pool.QueryRow(ctx,
		`SELECT
           (SELECT COUNT(*) FROM profile.profile_follows WHERE followee_id=$1),
           (SELECT COUNT(*) FROM profile.profile_follows WHERE follower_id=$1)`,
		userID).Scan(&followers, &following)
	return
}

func collectProfiles(rows pgx.Rows, queryErr error) ([]domain.Profile, error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()
	var out []domain.Profile
	for rows.Next() {
		var p domain.Profile
		if err := rows.Scan(&p.UserID, &p.Nickname, &p.DisplayName, &p.Bio, &p.AvatarMediaID,
			&p.City, &p.Country, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanProfile(row pgx.Row) (*domain.Profile, error) {
	var p domain.Profile
	if err := row.Scan(&p.UserID, &p.Nickname, &p.DisplayName, &p.Bio, &p.AvatarMediaID,
		&p.City, &p.Country, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}
