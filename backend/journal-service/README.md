# journal-service

Stores personal travel journal entries with rich text and attached media references.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /v1/journal/entries | X-User-Id header | List journal entries |
| POST | /v1/journal/entries | X-User-Id header | Create a journal entry |
| GET | /v1/journal/entries/:id | X-User-Id header | Get entry by ID |
| PATCH | /v1/journal/entries/:id | X-User-Id header | Update entry |
| DELETE | /v1/journal/entries/:id | X-User-Id header | Delete entry |
| GET | /v1/journal/entries/:id/media | X-User-Id header | List media attached to entry |
| POST | /v1/journal/entries/:id/media | X-User-Id header | Attach media to entry |
| DELETE | /v1/journal/entries/:id/media/:mid | X-User-Id header | Remove media from entry |
| GET | /healthz | — | Health check |

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| HTTP_ADDR | :8080 | Listen address |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_journal go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_journal go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up journal-service
```
