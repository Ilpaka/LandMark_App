# profile-service

Manages user profile data including display name, avatar, and privacy settings.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /v1/profile/me | X-User-Id header | Get own profile |
| PATCH | /v1/profile/me | X-User-Id header | Update own profile |
| GET | /v1/profile/:user_id | — | Get public profile by user ID |
| GET | /v1/profile/me/privacy | X-User-Id header | Get privacy settings |
| PATCH | /v1/profile/me/privacy | X-User-Id header | Update privacy settings |
| GET | /healthz | — | Health check |

> Auth is enforced at the API gateway; the service trusts the `X-User-Id` header injected by the gateway.

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| HTTP_ADDR | :8080 | Listen address |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_profile go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_profile go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up profile-service
```
