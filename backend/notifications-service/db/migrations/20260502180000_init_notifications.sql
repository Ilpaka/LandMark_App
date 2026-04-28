-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS notifications;
SET search_path TO notifications, public;

CREATE TABLE notifications_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    deep_link TEXT,
    meta JSONB NOT NULL DEFAULT '{}',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_notif_user_unread ON notifications_notifications(user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX ix_notif_user_all ON notifications_notifications(user_id, created_at DESC);

CREATE TABLE notifications_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('ios','android')),
    push_token TEXT NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_device_token ON notifications_devices(push_token);

CREATE TABLE notifications_preferences (
    user_id UUID PRIMARY KEY,
    push_trip_reminders BOOLEAN NOT NULL DEFAULT TRUE,
    push_moderation_result BOOLEAN NOT NULL DEFAULT TRUE,
    push_sync_status BOOLEAN NOT NULL DEFAULT FALSE,
    push_marketing BOOLEAN NOT NULL DEFAULT FALSE,
    email_welcome BOOLEAN NOT NULL DEFAULT TRUE,
    email_moderation_result BOOLEAN NOT NULL DEFAULT TRUE,
    email_marketing BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications_preferences;
DROP TABLE IF EXISTS notifications_devices;
DROP TABLE IF EXISTS notifications_notifications;
-- +goose StatementEnd
