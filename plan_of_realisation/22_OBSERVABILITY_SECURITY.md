# 22 · Observability и Security baseline

> Логи + метрики + трейсинг + базовый набор мер безопасности. Унаследовано из Tennis + расширения.

## 1. Logging

- Формат: **JSON, structured** через `slog.NewJSONHandler(os.Stdout, ...)`.
- Поля **обязательно** в каждой строке: `timestamp`, `level`, `service`, `msg`, `request_id`, `trace_id`, `span_id`.
- Чувствительные данные **не логируем**: пароли, OTP, refresh-токены, полные emails (маскируем `i**@example.com`). Маскирование — через `audit/mask.go` Tennis (копируется).
- Уровень из env `LOG_LEVEL`. По умолчанию `info`.

## 2. Metrics (Prometheus)

Каждый сервис экспортирует `GET /metrics`:

Стандартные метрики:
- `http_requests_total{service,method,path,status}` (counter)
- `http_request_duration_seconds{...}` (histogram, buckets дефолт)
- `db_pool_in_use`, `db_pool_idle` (gauge)
- `outbox_unpublished_total{service}` (gauge)
- `event_subscribe_lag_seconds{service,subject}` (gauge)

Бизнес-метрики (примеры):
- `auth_logins_total{result=success|fail}`
- `auth_register_total`
- `places_submitted_total`
- `moderation_decisions_total{action=approve|reject}`
- `media_uploads_finalized_total`
- `notifications_pushed_total{channel=push|email,result=ok|fail}`

## 3. Tracing (OpenTelemetry)

- SDK инициализируется в `cmd/server/main.go` через `pkg/observability.SetupOTel(ctx, serviceName, otlpEndpoint)`.
- Для Gin: автоматически через `otelgin.Middleware(serviceName)` (из Tennis).
- Кастомные спаны в use-case-уровне через `otel.Tracer(serviceName).Start(ctx, "<UseCase>")`.
- Контекст распространяется по HTTP-вызовам через стандартные `traceparent` headers (Gateway его пробрасывает).

В dev: spans отправляются в `otel-collector` → `tempo` → `grafana`.
В prod: на ваше предпочтение (Tempo, Datadog, Honeycomb…).

## 4. Health & readiness

- `GET /healthz` возвращает 200, если сервис может обработать запросы.
  - Проверяет: `SELECT 1` в Postgres (5s timeout), `PING` в Redis (3s).
  - Если не проходит — 503.
- В Docker `healthcheck:` секциях.

## 5. Security baseline

### 5.1 Secrets

- Все секреты — через env (никогда в коммитах).
- Production: использовать менеджер (Vault/Secrets Manager).
- `AUTH_PEPPER` ≥ 32 байт случайных.
- `INTERNAL_API_KEY` ≥ 32 байт.

### 5.2 Network

- Только api-gateway наружу. Все остальные сервисы — в Docker-internal сети.
- TLS-termination — на edge (Nginx/ALB) перед api-gateway.
- В docker-compose: `landmark-internal` driver=bridge, без `ports:` у downstream.

### 5.3 JWT

- RS256, 2048-bit. Ключи в `authkeys` томе.
- Access TTL 15м, Refresh TTL 30д.
- Refresh rotation + reuse detection (унаследовано Tennis).
- Blacklist by JTI в Redis (для logout).
- JWKS endpoint с rotation готов; в Phase 3 — реальный rotation (kid в заголовке + множественные ключи).

### 5.4 Passwords

- argon2id, memory=64 MiB, t=3, p=2 (унаследовано Tennis).
- Pepper применяется внутри hasher.
- Rate-limit на login: 5 попыток / 15 мин (унаследовано: lockout механизм).
- Минимальные требования пароля: ≥10 символов, минимум 1 буква + 1 цифра.

### 5.5 OTP

- 6-значный, TTL 15 мин.
- Сохраняем `salt + hash`, не plaintext.
- Cancellation: при создании нового — старые pending помечаются cancelled.

### 5.6 Input validation

- Все handlers используют binding-теги Gin (`required,email,uuid,...`).
- Доменные value-objects валидируются в use-cases.

### 5.7 SQL

- Только parametrized queries через `pgx`. Никаких string-concat.
- `pgcrypto` для `gen_random_uuid()`.

### 5.8 Mobile (Flutter)

- Hardening:
  - `flutter_secure_storage` для access/refresh — на iOS Keychain, на Android EncryptedSharedPreferences.
  - Сертификат-pinning в Dio (Phase 2): хеш SHA-256 кор-CA backend.
  - Disable debug-mode crashes в release builds.
  - Material `PinTextField` без отображения OTP в overlays.
  - На Android `android:allowBackup="false"` + `android:fullBackupContent="@xml/fullbackup_no_secure"` (только non-sensitive).
  - На iOS `NSAppTransportSecurity` строгий.

### 5.9 OWASP Mobile Top-10 чеклист

- [ ] M1: правильный TLS, без trust-all
- [ ] M2: secure local storage (no plain SharedPrefs для токенов)
- [ ] M3: secure communication (TLS 1.2+)
- [ ] M4: secure auth (refresh rotation)
- [ ] M5: cryptography (RS256 + argon2id)
- [ ] M6: secure authorization (JWT role)
- [ ] M7: code quality (lints + reviews)
- [ ] M8: code tampering detection (Phase 3 — Play Integrity / DeviceCheck)
- [ ] M9: reverse engineering resistance (Phase 3 — obfuscation)
- [ ] M10: extraneous functionality (отключаем `flutter inspector` в release)

## 6. Audit log

- Из Tennis: `auth_audit_log` для всех auth-событий + masked PII.
- В Phase 3 — добавим `places_audit_log`, `moderation_audit_log` для compliance.

## 7. Backups

- Postgres: `pg_dump` ежедневно в S3 (Phase 2 для prod).
- MinIO: replication / lifecycle policy.

## 8. SLOs (на Phase 2)

- Availability auth-service: 99.9% / 30 дней.
- API p95 < 300ms, p99 < 1s.
- Error budget: 0.1% (43 минуты в месяц).

## 9. Связанные файлы

- Полная инфра-настройка observability stack → `16_DOCKER_INFRA.md` §2 (профиль obs)
- Auth-internals → `05_AUTH_SERVICE.md`
