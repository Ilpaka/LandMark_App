-- +goose Up
-- +goose StatementBegin
SET search_path TO profile, public;

CREATE TABLE profile_follows (
    follower_id UUID NOT NULL REFERENCES profile_profiles(user_id) ON DELETE CASCADE,
    followee_id UUID NOT NULL REFERENCES profile_profiles(user_id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id <> followee_id)
);

CREATE INDEX ix_follows_followee ON profile_follows(followee_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profile.profile_follows;
-- +goose StatementEnd
