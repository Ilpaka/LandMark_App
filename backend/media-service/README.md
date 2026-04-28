# media-service

Handles media upload lifecycle via presigned MinIO/S3 URLs and tracks media metadata in Postgres.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /v1/media/uploads | X-User-Id header | Request a presigned PUT URL for upload |
| POST | /v1/media/uploads/:id/finalize | X-User-Id header | Mark upload as complete |
| GET | /v1/media/:id | X-User-Id header | Get media metadata |
| GET | /v1/media/:id/url | X-User-Id header | Get a presigned download URL |
| DELETE | /v1/media/:id | X-User-Id header | Delete media |
| GET | /healthz | — | Health check |

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| MINIO_ENDPOINT | minio:9000 | MinIO/S3 host:port |
| MINIO_ACCESS_KEY | — | Object storage access key |
| MINIO_SECRET_KEY | — | Object storage secret key |
| MINIO_BUCKET | landmark | Bucket name |
| MINIO_PUBLIC_ENDPOINT | — | Public-facing endpoint for URL generation |
| MINIO_USE_SSL | false | Use TLS for object storage connection |
| HTTP_ADDR | :8080 | Listen address |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_media go run ./cmd/migrate

# Start server (MinIO must be running)
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_media \
MINIO_ENDPOINT=localhost:9000 \
MINIO_ACCESS_KEY=minioadmin \
MINIO_SECRET_KEY=minioadmin \
go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up media-service
```
