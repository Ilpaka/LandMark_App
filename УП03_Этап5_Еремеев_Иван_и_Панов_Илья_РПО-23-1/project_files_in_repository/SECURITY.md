# Security Policy — LandMark (Wanderlog)

## Reporting
Found a vulnerability? Please open a private security advisory in the repository
or email the maintainers instead of filing a public issue.

## Secrets & configuration
- Real secrets live only in `.env.production` (or the deployment environment) and are
  **never committed**. `.gitignore` blocks `.env`, `.env.*`, `backups/`, dumps and logs;
  only `*.example` templates are tracked.
- `.env.production.example` ships placeholder values (`change_me_*`) only.
- `.env.demo.example` contains throwaway **demo** values for the local stand — not for production.

## Authentication & authorization
- Auth is JWT (RS256). The API Gateway validates the token and forwards trusted
  `X-User-Id` / `X-Role` headers; any client-supplied `X-Role` is stripped at the edge.
- Admin endpoints (`/v1/admin/*`) require `role=admin`; non-admins get `403`.
- Resource access is ownership-checked (a user cannot read/modify another user's data → `403`).

## Network exposure
- Only the API Gateway (`8080`) and operator UIs (Grafana `3000`, Prometheus `9090`,
  Mailpit `8025`, MinIO `9000/9001`) are published. Backend microservices, PostgreSQL
  (`5432`) and Redis (`6379`) are reachable only on the internal Docker network.

## Transport & CORS
- `CORS_ORIGINS` must be set to explicit origins in production (never `*`).
- TLS/HTTPS is expected to be terminated by the reverse proxy (see `deploy/nginx`).

## Backups
- PostgreSQL is backed up with `pg_dump` (`scripts/backup.bat`); restore is verified
  with `scripts/restore.bat`. Backups are stored outside version control.

## Dependencies
- Go modules are scanned with `govulncheck`; Dart deps with `flutter pub outdated`.
- Security patches (e.g. `golang-jwt/jwt/v5 >= 5.2.2`) are applied promptly.
