-- +goose Up
-- +goose StatementBegin
SET search_path TO places, public;

ALTER TABLE places_places
    ADD COLUMN search_vector tsvector
        GENERATED ALWAYS AS (
            to_tsvector('simple',
                coalesce(title, '') || ' ' ||
                coalesce(city, '') || ' ' ||
                coalesce(country, '') || ' ' ||
                coalesce(description, '')
            )
        ) STORED;

CREATE INDEX ix_places_fts ON places_places USING GIN(search_vector);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS places.ix_places_fts;
ALTER TABLE places.places_places DROP COLUMN IF EXISTS search_vector;
-- +goose StatementEnd
