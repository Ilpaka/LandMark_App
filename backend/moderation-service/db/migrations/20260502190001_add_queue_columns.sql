-- +goose Up
-- +goose StatementBegin
SET search_path TO moderation, public;

CREATE TABLE moderation.queue (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type  TEXT        NOT NULL,
    target_id    UUID        NOT NULL,
    submitted_by UUID        NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
    moderator_id UUID,
    note         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_queue_status_created ON moderation.queue(status, created_at DESC, id DESC) WHERE status = 'pending';
CREATE INDEX ix_queue_target ON moderation.queue(target_type, target_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS moderation.queue;
-- +goose StatementEnd
