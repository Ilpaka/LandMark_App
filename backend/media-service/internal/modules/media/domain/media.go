package domain

import (
	"github.com/google/uuid"
	"time"
)

type UploadStatus string

const (
	UploadPending   UploadStatus = "pending"
	UploadFinalized UploadStatus = "finalized"
	UploadExpired   UploadStatus = "expired"
)

type MediaStatus string

const (
	MediaReady   MediaStatus = "ready"
	MediaDeleted MediaStatus = "deleted"
)

type Upload struct {
	ID        uuid.UUID    `json:"id"`
	OwnerID   uuid.UUID    `json:"owner_id"`
	Kind      string       `json:"kind"`
	Mime      string       `json:"mime"`
	SizeBytes int64        `json:"size_bytes"`
	ObjectKey string       `json:"object_key"`
	ExpiresAt time.Time    `json:"expires_at"`
	Status    UploadStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
}

type Media struct {
	ID        uuid.UUID   `json:"id"`
	OwnerID   uuid.UUID   `json:"owner_id"`
	Kind      string      `json:"kind"`
	Mime      string      `json:"mime"`
	SizeBytes int64       `json:"size_bytes"`
	Width     int         `json:"width"`
	Height    int         `json:"height"`
	ObjectKey string      `json:"object_key"`
	Sha256    *string     `json:"sha256,omitempty"`
	Status    MediaStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type PresignedUpload struct {
	UploadID     uuid.UUID `json:"upload_id"`
	PresignedURL string    `json:"presigned_url"`
	ObjectKey    string    `json:"object_key"`
	ExpiresAt    time.Time `json:"expires_at"`
}
