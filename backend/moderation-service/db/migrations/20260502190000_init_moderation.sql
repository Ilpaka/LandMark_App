-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS moderation;
SET search_path TO moderation, public;

CREATE TABLE moderation_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id UUID NOT NULL UNIQUE,
    author_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','approved','rejected')),
    priority SMALLINT NOT NULL DEFAULT 0,
    assigned_to UUID,
    submitted_at TIMESTAMPTZ NOT NULL,
    decided_at TIMESTAMPTZ,
    decided_by UUID,
    decision TEXT CHECK (decision IS NULL OR decision IN ('approve','reject')),
    reason TEXT
);
CREATE INDEX ix_mod_open ON moderation_items(status, priority DESC, submitted_at ASC) WHERE status='open';

CREATE TABLE moderation_stats_daily (
    admin_id UUID NOT NULL,
    day DATE NOT NULL,
    approved INT NOT NULL DEFAULT 0,
    rejected INT NOT NULL DEFAULT 0,
    PRIMARY KEY (admin_id, day)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS moderation_stats_daily;
DROP TABLE IF EXISTS moderation_items;
-- +goose StatementEnd
