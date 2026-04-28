-- +goose Up
-- +goose StatementBegin
SET search_path TO journal, public;

CREATE TABLE journal_reactions (
    entry_id   UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    author_id  UUID NOT NULL,
    emoji      TEXT NOT NULL CHECK (char_length(emoji) <= 8),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (entry_id, author_id, emoji)
);
CREATE INDEX ix_reactions_entry ON journal_reactions(entry_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS journal.journal_reactions;
-- +goose StatementEnd
