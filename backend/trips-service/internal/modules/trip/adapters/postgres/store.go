package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Store { return &Store{db: db} }

const tripColumns = `id, owner_id, title, subtitle, start_date, end_date, cover_media_id,
    status, region, center_lat, center_lng, stops_count, entries_count, photos_count,
    created_at, updated_at`

func scanTrip(row pgx.Row, t *domain.Trip) error {
	return row.Scan(
		&t.ID, &t.OwnerID, &t.Title, &t.Subtitle, &t.StartDate, &t.EndDate, &t.CoverMediaID,
		&t.Status, &t.Region, &t.CenterLat, &t.CenterLng, &t.StopsCount, &t.EntriesCount, &t.PhotosCount,
		&t.CreatedAt, &t.UpdatedAt,
	)
}

func (s *Store) InsertTrip(ctx context.Context, t domain.Trip) (*domain.Trip, error) {
	err := scanTrip(s.db.QueryRow(ctx, fmt.Sprintf(`
        INSERT INTO trips_trips (id, owner_id, title, status, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6)
        RETURNING %s`, tripColumns),
		t.ID, t.OwnerID, t.Title, string(t.Status), t.CreatedAt, t.UpdatedAt,
	), &t)
	return &t, err
}

func (s *Store) GetTrip(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	var t domain.Trip
	err := scanTrip(s.db.QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM trips_trips WHERE id=$1`, tripColumns), id), &t)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) ListTrips(ctx context.Context, ownerID uuid.UUID, status, q, cursor string, limit int) ([]domain.Trip, string, error) {
	args := []any{ownerID}
	where := "owner_id = $1 AND status != 'archived'"
	argN := 2

	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, status)
		argN++
	}
	if q != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", argN)
		args = append(args, "%"+q+"%")
		argN++
	}

	// cursor-based pagination: cursor = base64(updated_at|id)
	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err == nil {
			parts := strings.SplitN(string(decoded), "|", 2)
			if len(parts) == 2 {
				where += fmt.Sprintf(" AND (updated_at, id) < ($%d::timestamptz, $%d::uuid)", argN, argN+1)
				args = append(args, parts[0], parts[1])
				argN += 2
			}
		}
	}

	args = append(args, limit+1)
	sqlQ := fmt.Sprintf(`SELECT %s FROM trips_trips WHERE %s ORDER BY updated_at DESC, id DESC LIMIT $%d`,
		tripColumns, where, argN)

	rows, err := s.db.Query(ctx, sqlQ, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var trips []domain.Trip
	for rows.Next() {
		var t domain.Trip
		if err := scanTrip(rows, &t); err != nil {
			return nil, "", err
		}
		trips = append(trips, t)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(trips) > limit {
		last := trips[limit-1]
		trips = trips[:limit]
		nextCursor = base64.StdEncoding.EncodeToString([]byte(
			last.UpdatedAt.UTC().Format(time.RFC3339Nano) + "|" + last.ID.String(),
		))
	}

	return trips, nextCursor, nil
}

func (s *Store) UpdateTrip(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.Trip, error) {
	sets := []string{"updated_at = now()"}
	args := []any{}
	argN := 1
	for k, v := range updates {
		switch k {
		case "title", "subtitle", "start_date", "end_date", "cover_media_id", "region", "center_lat", "center_lng":
			sets = append(sets, fmt.Sprintf("%s = $%d", k, argN))
			args = append(args, v)
			argN++
		}
	}
	args = append(args, id)
	var t domain.Trip
	err := scanTrip(s.db.QueryRow(ctx, fmt.Sprintf(
		`UPDATE trips_trips SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(sets, ", "), argN, tripColumns,
	), args...), &t)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &t, err
}

func (s *Store) SetTripStatus(ctx context.Context, id uuid.UUID, status domain.TripStatus) error {
	_, err := s.db.Exec(ctx, `UPDATE trips_trips SET status=$2, updated_at=now() WHERE id=$1`, id, string(status))
	return err
}

const stopColumns = `id, trip_id, place_id, title, latitude, longitude, planned_at, visited_at, notes, sort_order, created_at`

func scanStop(row pgx.Row, st *domain.TripStop) error {
	return row.Scan(
		&st.ID, &st.TripID, &st.PlaceID, &st.Title, &st.Latitude, &st.Longitude,
		&st.PlannedAt, &st.VisitedAt, &st.Notes, &st.SortOrder, &st.CreatedAt,
	)
}

func (s *Store) InsertStop(ctx context.Context, st domain.TripStop) (*domain.TripStop, error) {
	err := scanStop(s.db.QueryRow(ctx, fmt.Sprintf(`
        INSERT INTO trips_stops (id, trip_id, place_id, title, latitude, longitude, planned_at, notes, sort_order, created_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING %s`, stopColumns),
		st.ID, st.TripID, st.PlaceID, st.Title, st.Latitude, st.Longitude,
		st.PlannedAt, st.Notes, st.SortOrder, st.CreatedAt,
	), &st)
	return &st, err
}

func (s *Store) ListStops(ctx context.Context, tripID uuid.UUID) ([]domain.TripStop, error) {
	rows, err := s.db.Query(ctx,
		fmt.Sprintf(`SELECT %s FROM trips_stops WHERE trip_id=$1 ORDER BY sort_order, created_at`, stopColumns),
		tripID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stops []domain.TripStop
	for rows.Next() {
		var st domain.TripStop
		if err := scanStop(rows, &st); err != nil {
			return nil, err
		}
		stops = append(stops, st)
	}
	return stops, rows.Err()
}

func (s *Store) UpdateStop(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.TripStop, error) {
	sets := []string{}
	args := []any{}
	argN := 1
	for k, v := range updates {
		switch k {
		case "title", "place_id", "latitude", "longitude", "planned_at", "notes", "sort_order":
			sets = append(sets, fmt.Sprintf("%s = $%d", k, argN))
			args = append(args, v)
			argN++
		}
	}
	if len(sets) == 0 {
		return s.getStop(ctx, id)
	}
	args = append(args, id)
	var st domain.TripStop
	err := scanStop(s.db.QueryRow(ctx, fmt.Sprintf(
		`UPDATE trips_stops SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(sets, ", "), argN, stopColumns,
	), args...), &st)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &st, err
}

func (s *Store) getStop(ctx context.Context, id uuid.UUID) (*domain.TripStop, error) {
	var st domain.TripStop
	err := scanStop(s.db.QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM trips_stops WHERE id=$1`, stopColumns), id), &st)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &st, err
}

func (s *Store) DeleteStop(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM trips_stops WHERE id=$1`, id)
	return err
}

func (s *Store) MarkStopVisited(ctx context.Context, id uuid.UUID, at time.Time) error {
	_, err := s.db.Exec(ctx, `UPDATE trips_stops SET visited_at=$2 WHERE id=$1`, id, at)
	return err
}

func (s *Store) ListPublicTrips(ctx context.Context, excludeOwnerID uuid.UUID, cursor string, limit int) ([]domain.Trip, string, error) {
	where := "status IN ('planned','in_progress','completed') AND owner_id != $1"
	args := []any{excludeOwnerID}
	argN := 2

	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err == nil {
			parts := strings.SplitN(string(decoded), "|", 2)
			if len(parts) == 2 {
				where += fmt.Sprintf(" AND (updated_at, id) < ($%d::timestamptz, $%d::uuid)", argN, argN+1)
				args = append(args, parts[0], parts[1])
				argN += 2
			}
		}
	}

	args = append(args, limit+1)
	q := fmt.Sprintf(`SELECT %s FROM trips_trips WHERE %s ORDER BY updated_at DESC, id DESC LIMIT $%d`,
		tripColumns, where, argN)

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var trips []domain.Trip
	for rows.Next() {
		var t domain.Trip
		if err := scanTrip(rows, &t); err != nil {
			return nil, "", err
		}
		trips = append(trips, t)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(trips) > limit {
		last := trips[limit-1]
		trips = trips[:limit]
		nextCursor = base64.StdEncoding.EncodeToString([]byte(
			last.UpdatedAt.UTC().Format(time.RFC3339Nano) + "|" + last.ID.String(),
		))
	}
	return trips, nextCursor, nil
}
