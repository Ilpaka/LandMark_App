# api-gateway

Single entry point that validates JWTs, strips internal headers, applies CORS, and reverse-proxies requests to upstream services.

## Routing

| Prefix | Upstream env var | Default upstream |
|--------|-----------------|-----------------|
| /v1/auth, /v1/admin | AUTH_SERVICE_URL | http://auth-service:8080 |
| /v1/profile | PROFILE_SERVICE_URL | http://profile-service:8080 |
| /v1/places | PLACES_SERVICE_URL | http://places-service:8080 |
| /v1/trips | TRIPS_SERVICE_URL | http://trips-service:8080 |
| /v1/journal | JOURNAL_SERVICE_URL | http://journal-service:8080 |
| /v1/media | MEDIA_SERVICE_URL | http://media-service:8080 |
| /v1/favorites | FAVORITES_SERVICE_URL | http://favorites-service:8080 |
| /v1/notifications | NOTIFICATIONS_SERVICE_URL | http://notifications-service:8080 |
| /v1/moderation | MODERATION_SERVICE_URL | http://moderation-service:8080 |

Paths containing `/internal/` are blocked at the gateway. JWT is optional for `GET /v1/places` and all `/v1/auth/` public endpoints; all other routes require a valid Bearer token. Validated claims are forwarded as `X-User-Id`, `X-Role`, and `X-Session-Id` headers.

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| HTTP_ADDR | :8080 | Gateway listen address |
| AUTH_JWKS_URL | http://auth-service:8080/v1/auth/jwks | JWKS endpoint for JWT verification |
| JWT_ISSUER | landmark.app | Expected JWT iss claim |
| JWT_AUDIENCE | landmark-clients | Expected JWT aud claim |
| CORS_ORIGINS | * | Allowed CORS origins |
| AUTH_SERVICE_URL | http://auth-service:8080 | Auth service base URL |
| PROFILE_SERVICE_URL | http://profile-service:8080 | Profile service base URL |
| PLACES_SERVICE_URL | http://places-service:8080 | Places service base URL |
| TRIPS_SERVICE_URL | http://trips-service:8080 | Trips service base URL |
| JOURNAL_SERVICE_URL | http://journal-service:8080 | Journal service base URL |
| MEDIA_SERVICE_URL | http://media-service:8080 | Media service base URL |
| FAVORITES_SERVICE_URL | http://favorites-service:8080 | Favorites service base URL |
| NOTIFICATIONS_SERVICE_URL | http://notifications-service:8080 | Notifications service base URL |
| MODERATION_SERVICE_URL | http://moderation-service:8080 | Moderation service base URL |

## Running locally

```bash
# All upstream services must be reachable
AUTH_JWKS_URL=http://localhost:8081/.well-known/jwks.json \
AUTH_SERVICE_URL=http://localhost:8081 \
PROFILE_SERVICE_URL=http://localhost:8082 \
go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up api-gateway
```
