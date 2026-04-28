package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/media-service/internal/modules/media/domain"
	"github.com/ilpaka/landmark_app/backend/media-service/internal/modules/media/ports"
)

const (
	uploadTTL  = 15 * time.Minute
	uploadKeep = time.Hour
)

var allowedMimes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/webp": true, "image/heic": true,
}

type Service struct {
	Store  ports.Store
	S3     ports.ObjectStore
}

func (s *Service) RequestUpload(ctx context.Context, ownerID uuid.UUID, kind, mime string, size int64) (*domain.PresignedUpload, error) {
	if !allowedMimes[mime] {
		return nil, domain.ErrBadRequest
	}
	if kind != "photo" && kind != "avatar" {
		return nil, domain.ErrBadRequest
	}
	if size <= 0 || size > 20*1024*1024 {
		return nil, domain.ErrBadRequest
	}

	objectKey := fmt.Sprintf("%s/%s/%s", kind, ownerID, uuid.New())
	exp := time.Now().Add(uploadKeep)

	upload, err := s.Store.InsertUpload(ctx, domain.Upload{
		ID: uuid.New(), OwnerID: ownerID, Kind: kind, Mime: mime,
		SizeBytes: size, ObjectKey: objectKey, ExpiresAt: exp,
		Status: domain.UploadPending,
	})
	if err != nil {
		return nil, err
	}

	presignedURL, err := s.S3.PresignedPutURL(ctx, objectKey, mime, uploadTTL)
	if err != nil {
		return nil, err
	}

	return &domain.PresignedUpload{
		UploadID:     upload.ID,
		PresignedURL: presignedURL,
		ObjectKey:    objectKey,
		ExpiresAt:    time.Now().Add(uploadTTL),
	}, nil
}

func (s *Service) FinalizeUpload(ctx context.Context, uploadID, ownerID uuid.UUID) (*domain.Media, error) {
	upload, err := s.Store.GetUpload(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	if upload.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	if upload.Status != domain.UploadPending {
		return nil, domain.ErrBadRequest
	}

	if err := s.Store.FinalizeUpload(ctx, uploadID); err != nil {
		return nil, err
	}

	media, err := s.Store.InsertMedia(ctx, domain.Media{
		ID: uuid.New(), OwnerID: ownerID, Kind: upload.Kind,
		Mime: upload.Mime, SizeBytes: upload.SizeBytes,
		ObjectKey: upload.ObjectKey, Status: domain.MediaReady,
	})
	return media, err
}

func (s *Service) GetMedia(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	return s.Store.GetMedia(ctx, id)
}

func (s *Service) GetMediaURL(ctx context.Context, id uuid.UUID) (string, error) {
	m, err := s.Store.GetMedia(ctx, id)
	if err != nil {
		return "", err
	}
	return s.S3.PresignedGetURL(ctx, m.ObjectKey, 15*time.Minute)
}

func (s *Service) DeleteMedia(ctx context.Context, id, ownerID uuid.UUID) error {
	return s.Store.SoftDeleteMedia(ctx, id, ownerID)
}
