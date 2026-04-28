# favorites-service

Allows users to bookmark places, trips, and other entities, and organise them into named collections.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /v1/favorites | X-User-Id header | List all favorites |
| POST | /v1/favorites | X-User-Id header | Add a favorite |
| DELETE | /v1/favorites/:type/:id | X-User-Id header | Remove a favorite |
| GET | /v1/favorites/:type/:id | X-User-Id header | Check if item is favorited |
| GET | /v1/favorites/collections | X-User-Id header | List collections |
| POST | /v1/favorites/collections | X-User-Id header | Create a collection |
| PATCH | /v1/favorites/collections/:id | X-User-Id header | Update a collection |
| DELETE | /v1/favorites/collections/:id | X-User-Id header | Delete a collection |
| GET | /healthz | — | Health check |

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| HTTP_ADDR | :8080 | Listen address |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_favorites go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_favorites go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up favorites-service
```
