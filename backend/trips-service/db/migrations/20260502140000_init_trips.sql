-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS trips;
SET search_path TO trips, public;

CREATE TABLE trips_trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    title TEXT NOT NULL,
    subtitle TEXT,
    start_date DATE,
    end_date DATE,
    cover_media_id UUID,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','planned','in_progress','completed','archived')),
    region TEXT,
    center_lat FLOAT8,
    center_lng FLOAT8,
    stops_count INT NOT NULL DEFAULT 0,
    entries_count INT NOT NULL DEFAULT 0,
    photos_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);
CREATE INDEX ix_trips_owner ON trips_trips(owner_id, status, updated_at DESC);

CREATE TABLE trips_stops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips_trips(id) ON DELETE CASCADE,
    place_id UUID,
    title TEXT NOT NULL,
    latitude FLOAT8 NOT NULL,
    longitude FLOAT8 NOT NULL,
    planned_at DATE,
    visited_at TIMESTAMPTZ,
    notes TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_stops_trip ON trips_stops(trip_id, sort_order);

CREATE TABLE trips_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips_trips(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    day INT,
    color TEXT NOT NULL DEFAULT '#0E7C7B',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE trips_route_stops (
    route_id UUID NOT NULL REFERENCES trips_routes(id) ON DELETE CASCADE,
    stop_id UUID NOT NULL REFERENCES trips_stops(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    PRIMARY KEY (route_id, stop_id)
);

CREATE TABLE trips_outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trips_outbox;
DROP TABLE IF EXISTS trips_route_stops;
DROP TABLE IF EXISTS trips_routes;
DROP TABLE IF EXISTS trips_stops;
DROP TABLE IF EXISTS trips_trips;
-- +goose StatementEnd
