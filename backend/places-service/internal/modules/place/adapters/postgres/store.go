package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Store { return &Store{db: db} }

func (s *Store) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, slug, title, icon, color, sort_order, active, created_at
         FROM places_categories WHERE active = true ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Icon, &c.Color, &c.SortOrder, &c.Active, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (s *Store) InsertCategory(ctx context.Context, c domain.Category) (*domain.Category, error) {
	err := s.db.QueryRow(ctx,
		`INSERT INTO places_categories (id, slug, title, icon, color, sort_order, active, created_at)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, slug, title, icon, color, sort_order, active, created_at`,
		c.ID, c.Slug, c.Title, c.Icon, c.Color, c.SortOrder, c.Active, c.CreatedAt,
	).Scan(&c.ID, &c.Slug, &c.Title, &c.Icon, &c.Color, &c.SortOrder, &c.Active, &c.CreatedAt)
	return &c, err
}

func (s *Store) UpdateCategory(ctx context.Context, id uuid.UUID, title, icon, color string, sortOrder int) (*domain.Category, error) {
	var c domain.Category
	err := s.db.QueryRow(ctx,
		`UPDATE places_categories SET title=$2, icon=$3, color=$4, sort_order=$5
         WHERE id=$1 RETURNING id, slug, title, icon, color, sort_order, active, created_at`,
		id, title, icon, color, sortOrder,
	).Scan(&c.ID, &c.Slug, &c.Title, &c.Icon, &c.Color, &c.SortOrder, &c.Active, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &c, err
}

func (s *Store) DeactivateCategory(ctx context.Context, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `UPDATE places_categories SET active=false WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var c domain.Category
	err := s.db.QueryRow(ctx,
		`SELECT id, slug, title, icon, color, sort_order, active, created_at FROM places_categories WHERE slug=$1`,
		slug,
	).Scan(&c.ID, &c.Slug, &c.Title, &c.Icon, &c.Color, &c.SortOrder, &c.Active, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}

func (s *Store) ListPlaces(ctx context.Context, f domain.ListFilter) ([]domain.Place, error) {
	args := []any{}
	// Видимость:
	//   - публичные опубликованные — всем;
	//   - приватные — только их автору (любого статуса, обычно сразу published).
	where := "WHERE ((p.status = 'published' AND p.visibility = 'public')"
	argN := 1
	if f.ViewerID != nil {
		where += fmt.Sprintf(" OR (p.visibility = 'private' AND p.author_id = $%d)", argN)
		args = append(args, *f.ViewerID)
		argN++
	}
	where += ")"

	if f.BBox != nil {
		where += fmt.Sprintf(" AND p.latitude BETWEEN $%d AND $%d AND p.longitude BETWEEN $%d AND $%d",
			argN, argN+1, argN+2, argN+3)
		args = append(args, f.BBox.SouthLat, f.BBox.NorthLat, f.BBox.WestLng, f.BBox.EastLng)
		argN += 4
	}
	if f.Q != "" {
		// FTS via generated search_vector (migration 20260601000000_add_fts.sql);
		// also keep ILIKE so searches degrade gracefully before migration runs.
		where += fmt.Sprintf(` AND (p.title ILIKE $%d OR p.city ILIKE $%d OR
			p.search_vector @@ plainto_tsquery('simple', $%d))`, argN, argN, argN+1)
		args = append(args, "%"+f.Q+"%", f.Q)
		argN += 2
	}
	if f.CategorySlug != "" {
		where += fmt.Sprintf(` AND EXISTS(
            SELECT 1 FROM places_place_categories pc
            JOIN places_categories cat ON cat.id = pc.category_id
            WHERE pc.place_id = p.id AND cat.slug = $%d AND cat.active = true)`, argN)
		args = append(args, f.CategorySlug)
		argN++
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	args = append(args, limit)

	q := fmt.Sprintf(`
        SELECT p.id, p.title, p.description, p.latitude, p.longitude,
               p.address, p.city, p.country, p.author_id, p.status,
               p.reject_reason, p.source, p.visibility, p.cover_media_id, p.published_at, p.created_at, p.updated_at,
               COALESCE(
                 (SELECT array_agg(pc.category_id ORDER BY pc.category_id) FROM places_place_categories pc WHERE pc.place_id = p.id),
                 ARRAY[]::uuid[]
               ) AS category_ids
        FROM places_places p %s
        ORDER BY p.published_at DESC NULLS LAST
        LIMIT $%d`, where, argN)

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []domain.Place
	for rows.Next() {
		var p domain.Place
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Latitude, &p.Longitude,
			&p.Address, &p.City, &p.Country, &p.AuthorID, &p.Status,
			&p.RejectReason, &p.Source, &p.Visibility, &p.CoverMediaID, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
			&p.CategoryIDs,
		); err != nil {
			return nil, err
		}
		places = append(places, p)
	}
	return places, rows.Err()
}

func (s *Store) GetPlace(ctx context.Context, id uuid.UUID) (*domain.Place, error) {
	var p domain.Place
	err := s.db.QueryRow(ctx, `
        SELECT p.id, p.title, p.description, p.latitude, p.longitude,
               p.address, p.city, p.country, p.author_id, p.status,
               p.reject_reason, p.source, p.visibility, p.cover_media_id, p.published_at, p.created_at, p.updated_at,
               COALESCE(
                 (SELECT array_agg(pc.category_id ORDER BY pc.category_id) FROM places_place_categories pc WHERE pc.place_id = p.id),
                 ARRAY[]::uuid[]
               ) AS category_ids
        FROM places_places p WHERE p.id = $1`, id,
	).Scan(
		&p.ID, &p.Title, &p.Description, &p.Latitude, &p.Longitude,
		&p.Address, &p.City, &p.Country, &p.AuthorID, &p.Status,
		&p.RejectReason, &p.Source, &p.Visibility, &p.CoverMediaID, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
		&p.CategoryIDs,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) InsertPlace(ctx context.Context, p domain.Place) (*domain.Place, error) {
	if p.Visibility == "" {
		p.Visibility = domain.VisibilityPublic
	}
	publishedAt := p.PublishedAt
	err := s.db.QueryRow(ctx, `
        INSERT INTO places_places (id, title, description, latitude, longitude, address, city, country,
            author_id, status, source, visibility, cover_media_id, published_at, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
        RETURNING id, title, description, latitude, longitude, address, city, country,
                  author_id, status, reject_reason, source, visibility, cover_media_id, published_at, created_at, updated_at`,
		p.ID, p.Title, p.Description, p.Latitude, p.Longitude, p.Address, p.City, p.Country,
		p.AuthorID, string(p.Status), p.Source, string(p.Visibility), p.CoverMediaID, publishedAt, p.CreatedAt, p.UpdatedAt,
	).Scan(
		&p.ID, &p.Title, &p.Description, &p.Latitude, &p.Longitude,
		&p.Address, &p.City, &p.Country, &p.AuthorID, &p.Status,
		&p.RejectReason, &p.Source, &p.Visibility, &p.CoverMediaID, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	p.CategoryIDs = []uuid.UUID{}
	return &p, err
}

func (s *Store) UpdatePlace(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.Place, error) {
	sets := []string{"updated_at = now()"}
	args := []any{}
	argN := 1
	for k, v := range updates {
		switch k {
		case "title", "description", "address", "city", "country", "cover_media_id":
			sets = append(sets, fmt.Sprintf("%s = $%d", k, argN))
			args = append(args, v)
			argN++
		}
	}
	args = append(args, id)
	var p domain.Place
	err := s.db.QueryRow(ctx,
		fmt.Sprintf(`UPDATE places_places SET %s WHERE id = $%d
                     RETURNING id, title, description, latitude, longitude, address, city, country,
                               author_id, status, reject_reason, source, visibility, cover_media_id, published_at, created_at, updated_at`,
			strings.Join(sets, ", "), argN),
		args...,
	).Scan(
		&p.ID, &p.Title, &p.Description, &p.Latitude, &p.Longitude,
		&p.Address, &p.City, &p.Country, &p.AuthorID, &p.Status,
		&p.RejectReason, &p.Source, &p.Visibility, &p.CoverMediaID, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	p.CategoryIDs = []uuid.UUID{}
	return &p, err
}

func (s *Store) SetPlaceStatus(ctx context.Context, id uuid.UUID, status domain.PlaceStatus, extra map[string]any) error {
	q := `UPDATE places_places SET status=$2, updated_at=now()`
	args := []any{id, string(status)}
	argN := 3
	if extra != nil {
		if v, ok := extra["published_at"]; ok {
			q += fmt.Sprintf(", published_at=$%d", argN)
			args = append(args, v)
			argN++
		}
		if v, ok := extra["reject_reason"]; ok {
			q += fmt.Sprintf(", reject_reason=$%d", argN)
			args = append(args, v)
			argN++
		}
	}
	q += " WHERE id=$1"
	_, err := s.db.Exec(ctx, q, args...)
	return err
}

func (s *Store) SetPlaceCategories(ctx context.Context, placeID uuid.UUID, categoryIDs []uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM places_place_categories WHERE place_id=$1`, placeID)
	if err != nil {
		return err
	}
	for _, cid := range categoryIDs {
		if _, err := s.db.Exec(ctx, `INSERT INTO places_place_categories (place_id, category_id) VALUES ($1,$2)`, placeID, cid); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) InsertOutbox(ctx context.Context, aggType, evType string, payload map[string]any) error {
	b, _ := json.Marshal(payload)
	_, err := s.db.Exec(ctx, `INSERT INTO places_outbox (aggregate_type, event_type, payload) VALUES ($1,$2,$3)`, aggType, evType, b)
	return err
}
