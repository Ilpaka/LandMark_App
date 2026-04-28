package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/media-service/internal/modules/media/domain"
)

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) InsertUpload(ctx context.Context, u domain.Upload) (*domain.Upload, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO media.media_uploads(id,owner_id,kind,mime,size_bytes,object_key,expires_at,status)
         VALUES($1,$2,$3,$4,$5,$6,$7,$8)
         RETURNING id,owner_id,kind,mime,size_bytes,object_key,expires_at,status,created_at`,
		u.ID, u.OwnerID, u.Kind, u.Mime, u.SizeBytes, u.ObjectKey, u.ExpiresAt, string(u.Status))
	return scanUpload(row)
}

func (s *Store) GetUpload(ctx context.Context, id uuid.UUID) (*domain.Upload, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT id,owner_id,kind,mime,size_bytes,object_key,expires_at,status,created_at FROM media.media_uploads WHERE id=$1`, id)
	return scanUpload(row)
}

func (s *Store) FinalizeUpload(ctx context.Context, id uuid.UUID) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE media.media_uploads SET status='finalized' WHERE id=$1`, id)
	return err
}

func (s *Store) InsertMedia(ctx context.Context, m domain.Media) (*domain.Media, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO media.media_media(id,owner_id,kind,mime,size_bytes,width,height,object_key,sha256,status)
         VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
         RETURNING id,owner_id,kind,mime,size_bytes,width,height,object_key,sha256,status,created_at`,
		m.ID, m.OwnerID, m.Kind, m.Mime, m.SizeBytes, m.Width, m.Height, m.ObjectKey, m.Sha256, string(m.Status))
	return scanMedia(row)
}

func (s *Store) GetMedia(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT id,owner_id,kind,mime,size_bytes,width,height,object_key,sha256,status,created_at FROM media.media_media WHERE id=$1 AND status='ready'`, id)
	return scanMedia(row)
}

func (s *Store) SoftDeleteMedia(ctx context.Context, id, ownerID uuid.UUID) error {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE media.media_media SET status='deleted' WHERE id=$1 AND owner_id=$2 AND status='ready'`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanUpload(row pgx.Row) (*domain.Upload, error) {
	var u domain.Upload
	var st string
	if err := row.Scan(&u.ID, &u.OwnerID, &u.Kind, &u.Mime, &u.SizeBytes, &u.ObjectKey, &u.ExpiresAt, &st, &u.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u.Status = domain.UploadStatus(st)
	return &u, nil
}

func scanMedia(row pgx.Row) (*domain.Media, error) {
	var m domain.Media
	var st string
	if err := row.Scan(&m.ID, &m.OwnerID, &m.Kind, &m.Mime, &m.SizeBytes, &m.Width, &m.Height, &m.ObjectKey, &m.Sha256, &st, &m.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	m.Status = domain.MediaStatus(st)
	return &m, nil
}
