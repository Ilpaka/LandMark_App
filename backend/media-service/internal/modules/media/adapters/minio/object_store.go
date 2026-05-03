package miniostore

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStore struct {
	Client *minio.Client
	Bucket string
}

func New() (*ObjectStore, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "minio:9000"
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "landmark"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return nil, err
	}
	return &ObjectStore{Client: client, Bucket: bucket}, nil
}

func (s *ObjectStore) PresignedPutURL(ctx context.Context, objectKey, mime string, ttl time.Duration) (string, error) {
	params := url.Values{}
	params.Set("Content-Type", mime)
	u, err := s.Client.PresignedPutObject(ctx, s.Bucket, objectKey, ttl)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *ObjectStore) PresignedGetURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	u, err := s.Client.PresignedGetObject(ctx, s.Bucket, objectKey, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *ObjectStore) PublicURL(objectKey string) string {
	endpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("MINIO_ENDPOINT")
		if endpoint == "" {
			endpoint = "minio:9000"
		}
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "landmark"
	}
	return fmt.Sprintf("http://%s/%s/%s", endpoint, bucket, objectKey)
}
