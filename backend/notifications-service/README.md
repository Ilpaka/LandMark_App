# notifications-service

Delivers in-app notifications to users, manages push device tokens, and respects per-user preferences.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /v1/notifications | X-User-Id header | List notifications |
| POST | /v1/notifications/:id/read | X-User-Id header | Mark one notification as read |
| POST | /v1/notifications/read-all | X-User-Id header | Mark all notifications as read |
| POST | /v1/notifications/devices | X-User-Id header | Register a push device token |
| DELETE | /v1/notifications/devices/:id | X-User-Id header | Unregister a push device |
| GET | /v1/notifications/preferences | X-User-Id header | Get notification preferences |
| PATCH | /v1/notifications/preferences | X-User-Id header | Update notification preferences |
| POST | /v1/internal/notifications/send | X-Internal-Key | Send notification (internal only) |
| GET | /healthz | — | Health check |

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| HTTP_ADDR | :8080 | Listen address |
| INTERNAL_API_KEY | — | Shared key required on X-Internal-Key header for internal routes |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_notifications go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_notifications \
INTERNAL_API_KEY=secret \
go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up notifications-service
```
