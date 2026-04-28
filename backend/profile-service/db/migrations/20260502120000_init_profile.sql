-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS profile;
SET search_path TO profile, public;

CREATE TABLE profile_profiles (
    user_id UUID PRIMARY KEY,
    nickname TEXT NOT NULL,
    nickname_normalized TEXT NOT NULL,
    display_name TEXT NOT NULL,
    bio TEXT,
    avatar_media_id UUID,
    city TEXT,
    country TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_profile_nickname_norm ON profile_profiles(nickname_normalized);

CREATE TABLE profile_nickname_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname_normalized TEXT NOT NULL,
    user_id UUID,
    expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE profile_privacy_settings (
    user_id UUID PRIMARY KEY,
    profile_visibility TEXT NOT NULL DEFAULT 'public' CHECK (profile_visibility IN ('public','private')),
    trips_visibility TEXT NOT NULL DEFAULT 'private' CHECK (trips_visibility IN ('public','friends','private')),
    journal_visibility TEXT NOT NULL DEFAULT 'private' CHECK (journal_visibility IN ('friends','private')),
    analytics_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profile_privacy_settings;
DROP TABLE IF EXISTS profile_nickname_reservations;
DROP TABLE IF EXISTS profile_profiles;
-- +goose StatementEnd
