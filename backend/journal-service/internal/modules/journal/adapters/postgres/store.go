package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Store { return &Store{db: db} }

const entryColumns = `id, author_id, trip_id, place_id, latitude, longitude,
    title, body, mood, rating, occurred_at, created_at, updated_at, deleted_at`

func scanEntry(row pgx.Row, e *domain.Entry) error {
	return row.Scan(
		&e.ID, &e.AuthorID, &e.TripID, &e.PlaceID, &e.Latitude, &e.Longitude,
		&e.Title, &e.Body, &e.Mood, &e.Rating, &e.OccurredAt, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
}

func (s *Store) InsertEntry(ctx context.Context, e domain.Entry) (*domain.Entry, error) {
	err := scanEntry(s.db.QueryRow(ctx, fmt.Sprintf(`
        INSERT INTO journal_entries (id, author_id, trip_id, place_id, latitude, longitude,
            title, body, mood, rating, occurred_at, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
        RETURNING %s`, entryColumns),
		e.ID, e.AuthorID, e.TripID, e.PlaceID, e.Latitude, e.Longitude,
		e.Title, e.Body, e.Mood, e.Rating, e.OccurredAt, e.CreatedAt, e.UpdatedAt,
	), &e)
	if err != nil {
		return nil, err
	}
	e.MediaIDs = []uuid.UUID{}
	e.Tags = []string{}
	return &e, nil
}

func (s *Store) GetEntry(ctx context.Context, id uuid.UUID) (*domain.Entry, error) {
	var e domain.Entry
	err := scanEntry(s.db.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM journal_entries WHERE id=$1 AND deleted_at IS NULL`, entryColumns),
		id,
	), &e)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// load media and tags
	mediaIDs, err := s.ListMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	e.MediaIDs = mediaIDs

	tags, err := s.listTags(ctx, id)
	if err != nil {
		return nil, err
	}
	e.Tags = tags

	return &e, nil
}

func (s *Store) listTags(ctx context.Context, entryID uuid.UUID) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT tag FROM journal_entry_tags WHERE entry_id=$1 ORDER BY tag`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, rows.Err()
}

func (s *Store) ListEntries(ctx context.Context, authorID uuid.UUID, tripID *uuid.UUID, q, cursor string, limit int) ([]domain.Entry, string, error) {
	args := []any{authorID}
	where := "author_id = $1 AND deleted_at IS NULL"
	argN := 2

	if tripID != nil {
		where += fmt.Sprintf(" AND trip_id = $%d", argN)
		args = append(args, *tripID)
		argN++
	}
	if q != "" {
		where += fmt.Sprintf(" AND (title ILIKE $%d OR body ILIKE $%d)", argN, argN)
		args = append(args, "%"+q+"%")
		argN++
	}

	// cursor-based pagination by (occurred_at, id)
	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err == nil {
			parts := strings.SplitN(string(decoded), "|", 2)
			if len(parts) == 2 {
				where += fmt.Sprintf(" AND (occurred_at, id) < ($%d::timestamptz, $%d::uuid)", argN, argN+1)
				args = append(args, parts[0], parts[1])
				argN += 2
			}
		}
	}

	args = append(args, limit+1)
	sqlQ := fmt.Sprintf(`SELECT %s FROM journal_entries WHERE %s ORDER BY occurred_at DESC, id DESC LIMIT $%d`,
		entryColumns, where, argN)

	rows, err := s.db.Query(ctx, sqlQ, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var entries []domain.Entry
	for rows.Next() {
		var e domain.Entry
		if err := scanEntry(rows, &e); err != nil {
			return nil, "", err
		}
		e.MediaIDs = []uuid.UUID{}
		e.Tags = []string{}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(entries) > limit {
		last := entries[limit-1]
		entries = entries[:limit]
		nextCursor = base64.StdEncoding.EncodeToString([]byte(
			last.OccurredAt.UTC().Format(time.RFC3339Nano) + "|" + last.ID.String(),
		))
	}

	return entries, nextCursor, nil
}

func (s *Store) UpdateEntry(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.Entry, error) {
	sets := []string{"updated_at = now()"}
	args := []any{}
	argN := 1
	for k, v := range updates {
		switch k {
		case "title", "body", "mood", "rating", "occurred_at", "latitude", "longitude", "trip_id", "place_id":
			sets = append(sets, fmt.Sprintf("%s = $%d", k, argN))
			args = append(args, v)
			argN++
		}
	}
	args = append(args, id)
	var e domain.Entry
	err := scanEntry(s.db.QueryRow(ctx, fmt.Sprintf(
		`UPDATE journal_entries SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s`,
		strings.Join(sets, ", "), argN, entryColumns,
	), args...), &e)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	e.MediaIDs = []uuid.UUID{}
	e.Tags = []string{}
	return &e, nil
}

func (s *Store) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE journal_entries SET deleted_at=now() WHERE id=$1`, id)
	return err
}

func (s *Store) AddMedia(ctx context.Context, entryID, mediaID uuid.UUID, sortOrder int) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO journal_entry_media (entry_id, media_id, sort_order) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		entryID, mediaID, sortOrder,
	)
	return err
}

func (s *Store) RemoveMedia(ctx context.Context, entryID, mediaID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM journal_entry_media WHERE entry_id=$1 AND media_id=$2`, entryID, mediaID)
	return err
}

func (s *Store) ListMedia(ctx context.Context, entryID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.db.Query(ctx,
		`SELECT media_id FROM journal_entry_media WHERE entry_id=$1 ORDER BY sort_order`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if ids == nil {
		ids = []uuid.UUID{}
	}
	return ids, rows.Err()
}

func (s *Store) UpsertReaction(ctx context.Context, r domain.Reaction) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO journal_reactions (entry_id, author_id, emoji, created_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (entry_id, author_id, emoji) DO NOTHING`,
		r.EntryID, r.AuthorID, r.Emoji)
	return err
}

func (s *Store) DeleteReaction(ctx context.Context, entryID, authorID uuid.UUID, emoji string) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM journal_reactions WHERE entry_id=$1 AND author_id=$2 AND emoji=$3`,
		entryID, authorID, emoji)
	return err
}

func (s *Store) ListReactionCounts(ctx context.Context, entryID uuid.UUID) ([]domain.ReactionCount, error) {
	rows, err := s.db.Query(ctx,
		`SELECT emoji, count(*) FROM journal_reactions WHERE entry_id=$1 GROUP BY emoji ORDER BY count(*) DESC`,
		entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ReactionCount
	for rows.Next() {
		var rc domain.ReactionCount
		if err := rows.Scan(&rc.Emoji, &rc.Count); err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	if out == nil {
		out = []domain.ReactionCount{}
	}
	return out, rows.Err()
}

func (s *Store) GetUserReactions(ctx context.Context, entryID, authorID uuid.UUID) ([]string, error) {
	rows, err := s.db.Query(ctx,
		`SELECT emoji FROM journal_reactions WHERE entry_id=$1 AND author_id=$2 ORDER BY created_at`,
		entryID, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var emojis []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		emojis = append(emojis, e)
	}
	if emojis == nil {
		emojis = []string{}
	}
	return emojis, rows.Err()
}
