-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS media;
SET search_path TO media, public;

CREATE TABLE media_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('photo','avatar')),
    mime TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','finalized','expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_media_uploads_owner ON media_uploads(owner_id) WHERE status='pending';

CREATE TABLE media_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    kind TEXT NOT NULL,
    mime TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    object_key TEXT NOT NULL UNIQUE,
    sha256 TEXT,
    status TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('ready','deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_media_owner ON media_media(owner_id, created_at DESC) WHERE status='ready';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS media_media;
DROP TABLE IF EXISTS media_uploads;
-- +goose StatementEnd
