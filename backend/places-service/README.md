# places-service

Stores and serves landmark/place data including categories, with a moderation-gated approval workflow.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /v1/places/categories | — | List place categories |
| GET | /v1/places | — | List places (filterable) |
| GET | /v1/places/:id | — | Get place by ID |
| POST | /v1/places | X-User-Id header | Create a new place (draft) |
| PATCH | /v1/places/:id | X-User-Id header | Update a place |
| POST | /v1/places/:id/submit | X-User-Id header | Submit place for moderation |
| POST | /v1/places/internal/:id/approve | X-Internal-Key | Approve place (internal only) |
| POST | /v1/places/internal/:id/reject | X-Internal-Key | Reject place (internal only) |
| GET | /healthz | — | Health check |

> Internal endpoints are blocked at the API gateway and called only by moderation-service.

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| HTTP_ADDR | :8080 | Listen address |
| INTERNAL_API_KEY | dev-internal-key-change-me | Shared key for internal routes |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_places go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_places \
INTERNAL_API_KEY=secret \
go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up places-service
```
