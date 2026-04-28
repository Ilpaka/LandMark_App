-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS places;
SET search_path TO places, public;

CREATE TABLE places_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    icon TEXT NOT NULL,
    color TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE places_places (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    latitude FLOAT8 NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude FLOAT8 NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    address TEXT,
    city TEXT,
    country TEXT,
    author_id UUID,
    status TEXT NOT NULL CHECK (status IN ('draft','pending_moderation','published','rejected','archived')),
    reject_reason TEXT,
    source TEXT NOT NULL DEFAULT 'user' CHECK (source IN ('user','seed','imported')),
    cover_media_id UUID,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_places_status_pub ON places_places(status, published_at DESC) WHERE status='published';
CREATE INDEX ix_places_author ON places_places(author_id);
CREATE INDEX ix_places_geo ON places_places(latitude, longitude);

CREATE TABLE places_place_categories (
    place_id UUID NOT NULL REFERENCES places_places(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES places_categories(id) ON DELETE RESTRICT,
    PRIMARY KEY (place_id, category_id)
);

CREATE TABLE places_place_photos (
    place_id UUID NOT NULL REFERENCES places_places(id) ON DELETE CASCADE,
    media_id UUID NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    PRIMARY KEY (place_id, media_id)
);

CREATE TABLE places_outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);
CREATE INDEX ix_places_outbox_unpublished ON places_outbox(published_at) WHERE published_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS places_outbox;
DROP TABLE IF EXISTS places_place_photos;
DROP TABLE IF EXISTS places_place_categories;
DROP TABLE IF EXISTS places_places;
DROP TABLE IF EXISTS places_categories;
-- +goose StatementEnd
