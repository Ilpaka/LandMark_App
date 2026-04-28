-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS favorites;
SET search_path TO favorites, public;

CREATE TABLE favorites_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#0E7C7B',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_collection_default_per_user ON favorites_collections(user_id) WHERE is_default;

CREATE TABLE favorites_favorites (
    user_id UUID NOT NULL,
    target_type TEXT NOT NULL CHECK (target_type IN ('place','trip','entry')),
    target_id UUID NOT NULL,
    collection_id UUID,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, target_type, target_id)
);
CREATE INDEX ix_fav_user ON favorites_favorites(user_id, created_at DESC);
CREATE INDEX ix_fav_collection ON favorites_favorites(collection_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS favorites_favorites;
DROP TABLE IF EXISTS favorites_collections;
-- +goose StatementEnd
