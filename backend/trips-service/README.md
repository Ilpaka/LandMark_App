# trips-service

Manages travel trips and their ordered stops, including visit tracking.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /v1/trips | X-User-Id header | List user's trips |
| POST | /v1/trips | X-User-Id header | Create a trip |
| GET | /v1/trips/:id | X-User-Id header | Get trip by ID |
| PATCH | /v1/trips/:id | X-User-Id header | Update trip |
| DELETE | /v1/trips/:id | X-User-Id header | Delete trip |
| GET | /v1/trips/:id/stops | X-User-Id header | List stops for a trip |
| POST | /v1/trips/:id/stops | X-User-Id header | Add a stop to a trip |
| POST | /v1/trips/:id/stops/:sid/visit | X-User-Id header | Mark a stop as visited |
| GET | /healthz | — | Health check |

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| HTTP_ADDR | :8080 | Listen address |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_trips go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_trips go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up trips-service
```
