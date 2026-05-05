-- +goose Up
-- +goose StatementBegin
ALTER TABLE places_places
    ADD COLUMN visibility TEXT NOT NULL DEFAULT 'public'
        CHECK (visibility IN ('private','public'));

-- Author's own private places — direct lookup by (author_id, visibility).
CREATE INDEX ix_places_author_visibility
    ON places_places (author_id, visibility)
    WHERE visibility = 'private';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ix_places_author_visibility;
ALTER TABLE places_places DROP COLUMN IF EXISTS visibility;
-- +goose StatementEnd
