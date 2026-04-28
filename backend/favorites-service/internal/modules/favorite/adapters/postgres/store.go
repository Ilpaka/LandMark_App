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

	"github.com/ilpaka/landmark_app/backend/favorites-service/internal/modules/favorite/domain"
)

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) GetCollection(ctx context.Context, id uuid.UUID) (*domain.Collection, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT id,user_id,title,color,is_default,created_at FROM favorites.favorites_collections WHERE id=$1`, id)
	return scanCollection(row)
}

func (s *Store) GetDefaultCollection(ctx context.Context, userID uuid.UUID) (*domain.Collection, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT id,user_id,title,color,is_default,created_at FROM favorites.favorites_collections WHERE user_id=$1 AND is_default`, userID)
	return scanCollection(row)
}

func (s *Store) GetOrCreateDefaultCollection(ctx context.Context, userID uuid.UUID) (*domain.Collection, error) {
	col, err := s.GetDefaultCollection(ctx, userID)
	if err == nil {
		return col, nil
	}
	if err != pgx.ErrNoRows && !strings.Contains(err.Error(), "not found") {
		return nil, err
	}
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO favorites.favorites_collections(user_id,title,color,is_default)
         VALUES($1,'Избранное','#0E7C7B',TRUE)
         ON CONFLICT DO NOTHING
         RETURNING id,user_id,title,color,is_default,created_at`, userID)
	c, err := scanCollection(row)
	if err != nil {
		return s.GetDefaultCollection(ctx, userID)
	}
	return c, nil
}

func (s *Store) ListCollections(ctx context.Context, userID uuid.UUID) ([]domain.Collection, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id,user_id,title,color,is_default,created_at FROM favorites.favorites_collections WHERE user_id=$1 ORDER BY is_default DESC, created_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []domain.Collection
	for rows.Next() {
		c, err := scanCollection(rows)
		if err != nil {
			return nil, err
		}
		cols = append(cols, *c)
	}
	return cols, rows.Err()
}

func (s *Store) InsertCollection(ctx context.Context, userID uuid.UUID, title, color string) (*domain.Collection, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO favorites.favorites_collections(user_id,title,color) VALUES($1,$2,$3)
         RETURNING id,user_id,title,color,is_default,created_at`, userID, title, color)
	return scanCollection(row)
}

func (s *Store) UpdateCollection(ctx context.Context, id uuid.UUID, title, color string) (*domain.Collection, error) {
	row := s.Pool.QueryRow(ctx,
		`UPDATE favorites.favorites_collections SET title=$2,color=$3 WHERE id=$1
         RETURNING id,user_id,title,color,is_default,created_at`, id, title, color)
	return scanCollection(row)
}

func (s *Store) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM favorites.favorites_collections WHERE id=$1`, id)
	return err
}

func (s *Store) ListFavorites(ctx context.Context, f domain.ListFavoritesFilter) (*domain.FavoritesPage, error) {
	var conds []string
	args := []any{f.UserID}
	conds = append(conds, "user_id=$1")
	idx := 2

	if f.TargetType != nil {
		conds = append(conds, fmt.Sprintf("target_type=$%d", idx))
		args = append(args, string(*f.TargetType))
		idx++
	}
	if f.CollectionID != nil {
		conds = append(conds, fmt.Sprintf("collection_id=$%d", idx))
		args = append(args, *f.CollectionID)
		idx++
	}

	var cursorCond string
	if f.Cursor != "" {
		ts, id, err := decodeCursor(f.Cursor)
		if err == nil {
			cursorCond = fmt.Sprintf(" AND (created_at,target_id) < ($%d,$%d)", idx, idx+1)
			args = append(args, ts, id)
			idx += 2
		}
	}

	where := strings.Join(conds, " AND ")
	q := fmt.Sprintf(`SELECT user_id,target_type,target_id,collection_id,note,created_at FROM favorites.favorites_favorites WHERE %s%s ORDER BY created_at DESC, target_id DESC LIMIT $%d`,
		where, cursorCond, idx)
	args = append(args, f.Limit+1)

	rows, err := s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Favorite
	for rows.Next() {
		var fav domain.Favorite
		var tt string
		if err := rows.Scan(&fav.UserID, &tt, &fav.TargetID, &fav.CollectionID, &fav.Note, &fav.CreatedAt); err != nil {
			return nil, err
		}
		fav.TargetType = domain.TargetType(tt)
		items = append(items, fav)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	page := &domain.FavoritesPage{}
	if len(items) > f.Limit {
		last := items[f.Limit-1]
		page.NextCursor = encodeCursor(last.CreatedAt, last.TargetID)
		items = items[:f.Limit]
	}
	page.Items = items
	return page, nil
}

func (s *Store) GetFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID) (*domain.Favorite, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT user_id,target_type,target_id,collection_id,note,created_at FROM favorites.favorites_favorites WHERE user_id=$1 AND target_type=$2 AND target_id=$3`,
		userID, string(tt), targetID)
	var fav domain.Favorite
	var tts string
	if err := row.Scan(&fav.UserID, &tts, &fav.TargetID, &fav.CollectionID, &fav.Note, &fav.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	fav.TargetType = domain.TargetType(tts)
	return &fav, nil
}

func (s *Store) UpsertFavorite(ctx context.Context, fav domain.Favorite) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO favorites.favorites_favorites(user_id,target_type,target_id,collection_id,note)
         VALUES($1,$2,$3,$4,$5)
         ON CONFLICT(user_id,target_type,target_id) DO UPDATE SET collection_id=$4,note=$5`,
		fav.UserID, string(fav.TargetType), fav.TargetID, fav.CollectionID, fav.Note)
	return err
}

func (s *Store) DeleteFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID) error {
	_, err := s.Pool.Exec(ctx,
		`DELETE FROM favorites.favorites_favorites WHERE user_id=$1 AND target_type=$2 AND target_id=$3`,
		userID, string(tt), targetID)
	return err
}

func scanCollection(row pgx.Row) (*domain.Collection, error) {
	var c domain.Collection
	if err := row.Scan(&c.ID, &c.UserID, &c.Title, &c.Color, &c.IsDefault, &c.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
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
