-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS journal;
SET search_path TO journal, public;

CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL,
    trip_id UUID,
    place_id UUID,
    latitude FLOAT8,
    longitude FLOAT8,
    title TEXT,
    body TEXT NOT NULL DEFAULT '',
    mood TEXT CHECK (mood IS NULL OR mood IN ('happy','excited','calm','tired','sad')),
    rating SMALLINT CHECK (rating IS NULL OR rating BETWEEN 1 AND 5),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX ix_entries_author ON journal_entries(author_id, occurred_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX ix_entries_trip ON journal_entries(trip_id, occurred_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE journal_entry_media (
    entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    media_id UUID NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    PRIMARY KEY (entry_id, media_id)
);

CREATE TABLE journal_entry_tags (
    entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (entry_id, tag)
);

CREATE TABLE journal_outbox (
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
DROP TABLE IF EXISTS journal_outbox;
DROP TABLE IF EXISTS journal_entry_tags;
DROP TABLE IF EXISTS journal_entry_media;
DROP TABLE IF EXISTS journal_entries;
-- +goose StatementEnd
