# moderation-service

Queues user-submitted content for admin review and, on approval or rejection, calls back into places-service.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /v1/moderation/submit | X-User-Id header | Submit content for moderation |
| GET | /v1/moderation/queue | X-User-Id header (admin) | List pending moderation queue |
| POST | /v1/moderation/queue/:id/approve | X-User-Id header (admin) | Approve a queued item |
| POST | /v1/moderation/queue/:id/reject | X-User-Id header (admin) | Reject a queued item |
| POST | /v1/internal/moderation/submit | X-Internal-Key | Internal submit (service-to-service) |
| GET | /healthz | — | Health check |

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| HTTP_ADDR | :8080 | Listen address |
| INTERNAL_API_KEY | — | Shared key for X-Internal-Key header on internal routes |
| PLACES_SERVICE_URL | http://places-service:8080 | Internal URL of places-service for callbacks |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_moderation go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_moderation \
INTERNAL_API_KEY=secret \
PLACES_SERVICE_URL=http://localhost:8084 \
go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up moderation-service
```
