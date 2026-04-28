# auth-service

Handles user registration, login, JWT issuance, session management, OTP verification, and password reset.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /v1/auth/register | — | Register with email + password |
| POST | /v1/auth/login | — | Login, returns access + refresh tokens |
| POST | /v1/auth/verify-email | — | Confirm email OTP |
| POST | /v1/auth/refresh | — | Rotate access token using refresh token |
| POST | /v1/auth/logout | — | Invalidate current session |
| POST | /v1/auth/forgot-password | — | Send email reset OTP |
| POST | /v1/auth/reset-password | — | Apply email-based password reset |
| POST | /v1/auth/forgot-password/phone | — | Send SMS reset OTP |
| POST | /v1/auth/reset-password/phone | — | Apply phone-based password reset |
| POST | /v1/auth/phone/code | — | Request phone verification code |
| POST | /v1/auth/phone/verify | — | Verify phone OTP |
| POST | /v1/auth/phone/resend | — | Resend phone OTP |
| GET | /v1/auth/me | JWT | Current user info |
| POST | /v1/auth/me/email | JWT | Start email change |
| POST | /v1/auth/me/email/resend | JWT | Resend email verification |
| POST | /v1/auth/logout-all | JWT | Revoke all sessions |
| GET | /v1/auth/sessions | JWT | List active sessions |
| DELETE | /v1/auth/sessions/:id | JWT | Revoke specific session |
| POST | /v1/auth/change-password | JWT | Change password |
| DELETE | /v1/auth/account | JWT | Delete account |
| POST | /v1/admin/users/:id/block | JWT+admin | Block user |
| POST | /v1/admin/users/:id/unblock | JWT+admin | Unblock user |
| POST | /v1/admin/users/:id/force-logout | JWT+admin | Force logout user |
| GET | /v1/admin/users/:id/audit | JWT+admin | Audit log for user |
| GET | /.well-known/jwks.json | — | Public JWK keyset |
| GET | /healthz | — | Health check |
| GET | /metrics | — | Prometheus metrics |

## Environment Variables

| Var | Default | Description |
|-----|---------|-------------|
| DATABASE_URL | — | PostgreSQL connection string (required) |
| AUTH_PEPPER | — | Password hashing pepper (required) |
| JWT_PRIVATE_KEY_PATH | — | Path to RSA private key PEM (required) |
| JWT_PUBLIC_KEY_PATH | — | Path to RSA public key PEM (required) |
| REDIS_URL | redis://127.0.0.1:6379/0 | Redis connection URL |
| HTTP_ADDR | :8080 | Listen address |
| APP_ENV | development | Set to `production` for release mode |
| JWT_ISSUER | landmark.app | JWT iss claim |
| JWT_AUDIENCE | landmark-clients | JWT aud claim |
| PROFILE_SERVICE_URL | — | Internal URL of profile-service |
| ACCESS_TOKEN_TTL | 15m | Access token lifetime |
| REFRESH_TOKEN_TTL | 720h | Refresh token lifetime |
| SMSRU_API_ID | — | sms.ru API key for SMS OTP |
| OTEL_EXPORTER_OTLP_ENDPOINT | — | OTLP traces endpoint |

## Running locally

```bash
# Apply migrations
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_auth go run ./cmd/migrate

# Start server
DATABASE_URL=postgres://landmark:landmark@localhost:5432/landmark_auth \
AUTH_PEPPER=changeme \
JWT_PRIVATE_KEY_PATH=keys/private.pem \
JWT_PUBLIC_KEY_PATH=keys/public.pem \
go run ./cmd/server
```

## Docker

```bash
# From backend/ directory
docker compose up auth-service
```
