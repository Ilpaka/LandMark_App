package postgres

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/domain"
)

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) InsertItem(ctx context.Context, tt domain.TargetType, targetID, submittedBy uuid.UUID) (*domain.QueueItem, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO moderation.queue(target_type,target_id,submitted_by,status)
         VALUES($1,$2,$3,'pending')
         RETURNING id,target_type,target_id,submitted_by,status,moderator_id,note,created_at,updated_at`,
		string(tt), targetID, submittedBy)
	return scanItem(row)
}

func (s *Store) GetItem(ctx context.Context, id uuid.UUID) (*domain.QueueItem, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT id,target_type,target_id,submitted_by,status,moderator_id,note,created_at,updated_at FROM moderation.queue WHERE id=$1`, id)
	return scanItem(row)
}

func (s *Store) GetItemByTarget(ctx context.Context, tt domain.TargetType, targetID uuid.UUID) (*domain.QueueItem, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT id,target_type,target_id,submitted_by,status,moderator_id,note,created_at,updated_at FROM moderation.queue WHERE target_type=$1 AND target_id=$2 ORDER BY created_at DESC LIMIT 1`,
		string(tt), targetID)
	return scanItem(row)
}

func (s *Store) ListPending(ctx context.Context, f domain.ListFilter) (*domain.QueuePage, error) {
	args := []any{"pending"}
	q := `SELECT id,target_type,target_id,submitted_by,status,moderator_id,note,created_at,updated_at FROM moderation.queue WHERE status=$1`
	idx := 2

	if f.Cursor != "" {
		ts, id, err := decodeCursor(f.Cursor)
		if err == nil {
			q += fmt.Sprintf(" AND (created_at,id) < ($%d,$%d)", idx, idx+1)
			args = append(args, ts, id)
			idx += 2
		}
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", idx)
	args = append(args, f.Limit+1)

	rows, err := s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.QueueItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	page := &domain.QueuePage{}
	if len(items) > f.Limit {
		last := items[f.Limit-1]
		page.NextCursor = encodeCursor(last.CreatedAt, last.ID)
		items = items[:f.Limit]
	}
	page.Items = items
	return page, nil
}

func (s *Store) UpdateDecision(ctx context.Context, id uuid.UUID, status domain.ItemStatus, moderatorID uuid.UUID, note *string) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE moderation.queue SET status=$2,moderator_id=$3,note=$4,updated_at=now() WHERE id=$1`,
		id, string(status), moderatorID, note)
	return err
}

func (s *Store) GetStats(ctx context.Context) (*domain.ModerationStats, error) {
	var stats domain.ModerationStats
	err := s.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status='pending'),
			COUNT(*) FILTER (WHERE status='approved'),
			COUNT(*) FILTER (WHERE status='rejected'),
			COUNT(*)
		FROM moderation.queue
	`).Scan(&stats.Pending, &stats.Approved, &stats.Rejected, &stats.Total)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func scanItem(row pgx.Row) (*domain.QueueItem, error) {
	var item domain.QueueItem
	var tt, st string
	if err := row.Scan(&item.ID, &tt, &item.TargetID, &item.SubmittedBy, &st, &item.ModeratorID, &item.Note, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	item.TargetType = domain.TargetType(tt)
	item.Status = domain.ItemStatus(st)
	return &item, nil
}

func encodeCursor(t time.Time, id uuid.UUID) string {
	s := fmt.Sprintf("%s|%s", t.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func decodeCursor(c string) (time.Time, uuid.UUID, error) {
	b, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("bad cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	id, err := uuid.Parse(parts[1])
	return t, id, err
}
