# 05 · auth-service: копирование Tennis и адаптация

> Это **самый важный** файл бэкенда: он подробно описывает, как ИИ должен скопировать `Tennis_app/backend/auth-service` в `LandMark_App/backend/auth-service` и адаптировать под Wanderlog.

## 1. Источник

```
/Users/ilpaka/Development/Tennis_app/backend/auth-service/
```

Содержит готовую, протестированную реализацию:
- Email + пароль регистрация и подтверждение через 6-значный OTP
- Phone + SMS OTP логин (включая регистрацию по номеру)
- Login по email+пароль
- Refresh-токены с rotation + reuse detection
- Sessions list, revoke, revoke-all
- Forgot password (email и phone-варианты)
- Change password
- Delete account
- Admin: block/unblock/force-logout/audit
- JWKS endpoint
- Audit log + transactional outbox
- OpenAPI 3.0 контракт + schemathesis-фьюзинг

Эту реализацию **сохраняем целиком**, меняем только модульный путь и убираем бизнес-специфичный «теннис» (которого там почти нет — только в названиях БД-юзера и сервиса).

## 2. План копирования (пошагово)

### Шаг 1. Скопировать общий пакет observability

```sh
mkdir -p /Users/ilpaka/Development/LandMark_App/backend/pkg
cp -R /Users/ilpaka/Development/Tennis_app/backend/pkg/observability \
      /Users/ilpaka/Development/LandMark_App/backend/pkg/observability
```

В `backend/pkg/observability/go.mod` поменять module path:
- `github.com/ilpaka/tennis_app/backend/pkg/observability` → `github.com/ilpaka/landmark_app/backend/pkg/observability`

### Шаг 2. Скопировать сам auth-service

```sh
cp -R /Users/ilpaka/Development/Tennis_app/backend/auth-service \
      /Users/ilpaka/Development/LandMark_App/backend/auth-service
```

Удалить артефакты предыдущих сборок:
```sh
rm -f /Users/ilpaka/Development/LandMark_App/backend/auth-service/coverage.out
rm -rf /Users/ilpaka/Development/LandMark_App/backend/auth-service/keys/*
```

### Шаг 3. Поменять module path во всех файлах

В `backend/auth-service/go.mod`:
```
module github.com/ilpaka/tennis_app/backend/auth-service  →
module github.com/ilpaka/landmark_app/backend/auth-service
```

И replace-директива:
```
replace github.com/ilpaka/tennis_app/backend/pkg/observability => ../pkg/observability  →
replace github.com/ilpaka/landmark_app/backend/pkg/observability => ../pkg/observability
```

Затем во ВСЕХ Go-файлах рекурсивно:
```sh
cd /Users/ilpaka/Development/LandMark_App/backend/auth-service
grep -rl 'github.com/ilpaka/tennis_app' . | xargs sed -i '' 's|github.com/ilpaka/tennis_app|github.com/ilpaka/landmark_app|g'
```

После этого: `go mod tidy`.

### Шаг 4. Конфигурация — переименование сервиса в логах/трейсах

В `internal/config/config.go`:
- `OTEL_SERVICE_NAME` дефолт: `tennis-auth-api` → `landmark-auth-api`.

В `cmd/server/main.go` и других местах поиск по `tennis` (если найдётся) — заменить на `landmark`.

### Шаг 5. Имена БД-юзера и БД

В `Tennis_app` `docker-compose.yml`:
```yaml
POSTGRES_USER: tennis
POSTGRES_PASSWORD: tennis
POSTGRES_DB: tennis
```

В **корневом** docker-compose `LandMark_App` (см. `16_DOCKER_INFRA.md`) запустим Postgres под `landmark/landmark/landmark`. Соответственно, env auth-service:
```
DATABASE_URL=postgres://landmark:landmark@postgres:5432/landmark?search_path=auth
```

То есть БД одна (`landmark`), а каждый сервис подключается с `search_path` под свою схему. Миграции auth уже создают свои таблицы (`auth_accounts`, `auth_credentials`, …), но **не в схеме auth, а в public**. Поэтому нужно адаптировать миграции:

### Шаг 6. Адаптация миграций — перенос в схему `auth`

В **первой** миграции (`db/migrations/20260117120000_auth_schema.sql`) добавить в начало:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS auth;
SET search_path TO auth, public;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE auth_accounts ( ... );  -- остальные таблицы как есть
...
-- +goose StatementEnd
```

> Альтернатива — оставить всё в `public` и не использовать схему. Это допустимо в MVP, но мы выберем схемы ради изоляции (см. `03_ARCHITECTURE.md` §6.1).

И в `cmd/migrate/main.go` убедиться, что подключение использует `?search_path=auth,public` либо явный `SET search_path` в коде. Самый простой путь — пропустить через env уже готовый DSN.

### Шаг 7. Имя сервиса в OpenAPI

`api/openapi/auth.yaml`:
```yaml
info:
  title: Tennis Auth API → LandMark Auth API
```

### Шаг 8. Подгонка под нашу архитектуру: email-сразу-регистрация

Tennis сейчас **не имеет** прямого эндпоинта `POST /v1/auth/register` (email+пароль одной формой). Вместо этого там phone-flow создаёт аккаунт. Нам для Wanderlog нужно так:

1. `POST /v1/auth/register` body: `{name, email, password}` →
   - Создаёт `auth_accounts` со `status=pending_verification`, `email`, `email_normalized`.
   - Создаёт `auth_credentials` с argon2id-хешем пароля.
   - Создаёт запись в `auth_email_verifications`, отправляет 6-значный OTP на email через `EmailSender`.
   - Резервирует никнейм через `ProfileClient` (можно отложить до `verify-email`).
   - Возвращает `{verification_id}`.
2. `POST /v1/auth/verify-email` body: `{verification_id, code}` →
   - Помечает email_verified, account.status=`active`.
   - Подтверждает резервацию никнейма в profile-service.
   - Записывает в outbox: `auth.user_registered`.
3. После этого пользователь может `POST /v1/auth/login` (этот эндпоинт уже есть).

**Что добавить в код Tennis:**

В `internal/modules/auth/app/service_ops.go` (или новом файле `register.go`) — метод `RegisterEmail(ctx, email, password, name) (verificationID uuid.UUID, err error)`:

```go
func (s *Service) RegisterEmail(ctx context.Context, email, password, name string) (uuid.UUID, error) {
    if !validPassword(password) || name == "" { return uuid.Nil, domain.ErrBadRequest }
    addr, norm, err := normalizeEmail(email); if err != nil { return uuid.Nil, err }

    var verID uuid.UUID
    err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
        // 1. Проверить, что нет активного аккаунта с этим email
        if existing, _ := s.Store.GetAccountByEmailNorm(ctx, tx, norm); existing != nil &&
            existing.Status != domain.AccountStatusDeleted { return domain.ErrConflict }

        // 2. Создать аккаунт
        acc, err := s.Store.InsertAccount(ctx, tx, addr, norm,
            domain.AccountStatusPendingVerification, domain.RoleUser)
        if err != nil { return err }

        // 3. Захэшировать пароль (Pepper применяется внутри hasher)
        hash, err := s.Hasher.Hash(password + s.Pepper); if err != nil { return err }
        if err := s.Store.UpsertCredential(ctx, tx, acc.ID, hash, 1); err != nil { return err }

        // 4. Создать verification + отправить OTP (через outbox или прямой вызов)
        salt := crypto.RandomBytes(16)
        code := crypto.GenerateOTP6()
        codeHash := crypto.HashOTP(salt, code)
        v, err := s.Store.InsertVerification(ctx, tx, acc.ID, addr, salt, codeHash,
            time.Now().Add(s.VerifyTTL))
        if err != nil { return err }

        // 5. Резервируем никнейм через ProfileClient (по name)
        // (опц., можно вынести в `verify-email`)
        // _, err := s.Profile.ReserveNickname(ctx, slug(name))

        // 6. Запись в outbox для отправки welcome-почты после verify
        _ = s.Store.InsertOutbox(ctx, tx, ports.OutboxMessage{
            AggregateType: "auth_account",
            EventType:     "registration_started",
            Payload:       map[string]any{"user_id": acc.ID, "email": norm},
        })

        verID = v.ID
        // Отправить OTP синхронно (упрощение MVP)
        go s.Mail.SendVerificationOTP(context.Background(), addr, code) // detach
        return nil
    })
    return verID, err
}
```

Зарегистрировать в роутере:
```go
r.POST("/auth/register", h.Register)
```

И handler:
```go
type registerReq struct {
    Name     string `json:"name" binding:"required,min=1,max=64"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=10,max=128"`
}

func (h *Handlers) Register(c *gin.Context) {
    var req registerReq
    if err := c.ShouldBindJSON(&req); err != nil { writeErr(c, domain.ErrBadRequest); return }
    if !validPassword(req.Password) { writeErr(c, domain.ErrBadRequest); return }
    vid, err := h.Svc.RegisterEmail(c.Request.Context(), req.Email, req.Password, req.Name)
    if err != nil { writeErr(c, err); return }
    c.JSON(http.StatusOK, gin.H{"status": "ok", "verification_id": vid.String()})
}
```

Соответственно обновить OpenAPI (`api/openapi/auth.yaml`) — добавить `/v1/auth/register` с RegistrationRequest и теми же error-кодами, что и phone-flow.

### Шаг 9. Отключение phone-flow в MVP

Phone-flow в Tennis уже работает, но в нашем MVP мы его **скрываем в UI** через feature flag `feature.phoneAuth=false`. На бэке оставляем эндпоинты `/v1/auth/phone/*` нетронутыми (их использование никому не мешает).

### Шаг 10. Адаптация ProfileClient

Tennis включает `ProfileClient` для резервирования никнейма. У нас это будет `profile-service` (см. `06_PROFILE_SERVICE.md`). Эндпоинты совпадают:

```
POST /v1/profile/internal/nicknames/reserve  body: {nickname} → {reservation_id}
POST /v1/profile/internal/nicknames/{rid}/confirm  body: {user_id}
DELETE /v1/profile/internal/nicknames/{rid}
```

Эндпоинт `/internal/*` помечен как **service-to-service**: api-gateway его не проксирует, доступ только из Docker-сети. Аутентификация через shared secret в env: `INTERNAL_API_KEY`.

В `internal/modules/auth/adapters/profile/client.go` (унаследовано) поменять только `BASE URL` через env `PROFILE_SERVICE_URL`.

### Шаг 11. Конкретное изменение в OpenAPI (полный список новых/изменённых эндпоинтов)

| Метод | Путь | Источник | Действие |
|---|---|---|---|
| POST | /v1/auth/register | новый | добавить (см. Шаг 8) |
| POST | /v1/auth/verify-email | есть | без изменений |
| POST | /v1/auth/login | есть | без изменений |
| POST | /v1/auth/refresh | есть | без изменений |
| POST | /v1/auth/logout | есть | без изменений |
| GET | /v1/auth/me | есть | без изменений |
| GET | /v1/auth/sessions | есть | без изменений |
| DELETE | /v1/auth/sessions/{id} | есть | без изменений |
| POST | /v1/auth/me/email | есть | без изменений |
| POST | /v1/auth/me/email/resend | есть | без изменений |
| POST | /v1/auth/change-password | есть | без изменений |
| POST | /v1/auth/forgot-password | есть | без изменений |
| POST | /v1/auth/reset-password | есть | без изменений |
| DELETE | /v1/auth/account | есть | без изменений |
| GET | /v1/auth/jwks | есть | без изменений |
| POST | /v1/auth/phone/* | есть | оставить, но скрыть в UI |
| POST | /v1/admin/users/{id}/block | есть | без изменений |
| POST | /v1/admin/users/{id}/unblock | есть | без изменений |
| POST | /v1/admin/users/{id}/force-logout | есть | без изменений |
| GET | /v1/admin/users/{id}/audit | есть | без изменений |

## 3. Структура каталогов после копирования

```
backend/auth-service/
├── api/openapi/auth.yaml
├── cmd/{server,migrate}/main.go
├── db/migrations/*.sql           # +1 префикс схемы
├── internal/
│   ├── config/config.go
│   ├── httpserver/{server,middleware,swagger}.go
│   └── modules/auth/
│       ├── domain/{account,errors,phone_otp}.go
│       ├── ports/{ports,generate}.go
│       ├── app/{service,service_ops,me,phone_flow,phone_reset_flow,session_tokens,
│       │       audit_enqueue,email_settings,ratelimit,register}.go  # +register.go
│       ├── adapters/
│       │   ├── jwt/signer.go
│       │   ├── postgres/store.go
│       │   ├── redis/{client,phone_otp}.go
│       │   ├── sms/{log,smsru}.go
│       │   ├── email/sender.go
│       │   └── profile/client.go
│       ├── http/{router,handlers,errors,headers,middleware}.go
│       ├── worker/outbox.go
│       ├── metrics/metrics.go
│       ├── crypto/{tokens,otp,password}.go
│       ├── audit/{mask,payload}.go
│       └── mocks/mock_store.go
├── tests/{e2e,integration,security,testutil}/
├── scripts/schemathesis.sh
├── docker-entrypoint.sh
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

## 4. Переменные окружения auth-service

Полный список (унаследован от Tennis + добавки):

| Env | Default | Назначение |
|---|---|---|
| APP_ENV | `development` | `production` включает gin release-mode |
| HTTP_ADDR | `:8080` | Адрес сервера |
| DATABASE_URL | — (req) | Postgres DSN с `?search_path=auth,public` |
| REDIS_URL | `redis://127.0.0.1:6379/0` | Redis DSN |
| JWT_PRIVATE_KEY_PATH | `/app/keys/jwt_rsa.pem` | RSA private (генерится в entrypoint) |
| JWT_PUBLIC_KEY_PATH | `/app/keys/jwt_rsa.pub.pem` | RSA public |
| JWT_ISSUER | `landmark.app` | iss claim |
| JWT_AUDIENCE | `landmark-clients` | aud claim |
| AUTH_PEPPER | — (req) | Pepper для argon2id, **секрет**, не коммитится |
| PROFILE_SERVICE_URL | `http://profile-service:8080` | URL для резервации никнейма |
| INTERNAL_API_KEY | — (req) | Shared secret для service-to-service вызовов |
| ACCESS_TOKEN_TTL | `15m` | TTL access JWT |
| REFRESH_TOKEN_TTL | `720h` (30d) | TTL refresh |
| VERIFICATION_OTP_TTL | `15m` | TTL OTP подтверждения email |
| RESET_OTP_TTL | `15m` | TTL OTP сброса пароля |
| AUTH_ARGON_MEMORY_KIB | `65536` | Argon2id memory cost |
| AUTH_ARGON_TIME | `3` | Argon2id iterations |
| AUTH_ARGON_PARALLELISM | `2` | Argon2id parallelism |
| OTEL_SERVICE_NAME | `landmark-auth-api` | OTel resource |
| OTEL_EXPORTER_OTLP_ENDPOINT | (empty) | Если задан — отправляем спаны |
| SMSRU_API_ID | (empty) | Реальный SMS-провайдер (опционально) |
| SMTP_HOST/PORT/USER/PASS/FROM | local Mailhog | SMTP настройки (см. файл 12) |

## 5. Тесты, которые должны зелёным запуститься после копирования

Из Tennis уже есть:
- `internal/modules/auth/app/service_test.go` — service-level
- `internal/modules/auth/app/ratelimit_test.go`
- `internal/modules/auth/http/handlers_test.go`, `handlers_auth_test.go`, `errors_test.go`
- `internal/modules/auth/crypto/crypto_test.go`
- `internal/modules/auth/adapters/jwt/signer_test.go`
- `internal/modules/auth/adapters/postgres/store_integration_test.go` (build tag integration)
- `internal/modules/auth/adapters/redis/{client,phone_otp}_test.go`
- `internal/modules/auth/audit/mask_test.go`
- `tests/e2e/...` (build tag e2e)
- `tests/integration/...` (build tag integration)
- `tests/security/...`

После замены module path и реорганизации схемы все тесты должны зеленеть. Дополнительно написать:
- `internal/modules/auth/app/register_test.go` — для нового RegisterEmail.
- e2e: full email registration flow (`/register` → `/verify-email` → `/login` → `/me`).

Минимальное coverage по `make cover-check` — `MIN_COVER=12` оставляем (можем поднять до 30 в Phase 2).

## 6. JWKS endpoint и downstream-сервисы

auth-service выставляет `GET /v1/auth/jwks` (унаследовано) — JSON Web Key Set с публичным ключом. Все остальные сервисы в `pkg/authmid` (см. файл 04) кэшируют JWKS на 1 час и используют для валидации Access JWT.

## 7. Audit log: что писать

Унаследовано из Tennis. Кратко: на каждое значимое действие (`account_registered`, `email_verified`, `login_success`, `login_failed`, `password_changed`, `password_reset`, `session_revoked`, `account_blocked`, `account_deleted`, `admin_*`) пишем строку в `auth_audit_log` (с маскированием PII через `audit/mask.go`).

Этот лог используется в админке через `GET /v1/admin/users/{id}/audit`.

## 8. Связь с client-side (Flutter)

Flutter-фича `features/auth/` использует следующие client-side сервисы:
- `AuthApiClient` (Dio + interceptor для refresh) — описан в `20_FLUTTER_FEATURES.md`.
- Хранение токенов: `flutter_secure_storage` ключ `auth.access`, `auth.refresh`.
- При `401` interceptor автоматически вызывает `/v1/auth/refresh`. Если refresh неуспешен — `logout()` и переход на `AU-01`.

## 9. Чек-лист завершения копирования

- [ ] `go test ./...` проходит без ошибок
- [ ] `make test-integration` проходит (требует docker для dockertest)
- [ ] `docker compose up auth-service` поднимается, миграции применяются
- [ ] `curl http://localhost:8080/v1/auth/jwks` возвращает JSON
- [ ] `curl -X POST http://localhost:8080/v1/auth/register -d '{"name":"X","email":"x@y.z","password":"Aa12345678"}'` возвращает `verification_id`
- [ ] `OPENAPI_AUTH_YAML=/app/api/openapi/auth.yaml` валидируется через `kin-openapi` без ошибок
- [ ] OTel-spans видны в Tempo/Jaeger при `OTEL_EXPORTER_OTLP_ENDPOINT`

## 10. Связанные файлы

- Профиль-сервис (резервирование никнеймов) → `06_PROFILE_SERVICE.md`
- API Gateway → `14_API_GATEWAY.md`
- Docker-инфра → `16_DOCKER_INFRA.md`
- Контракт → `17_API_CONTRACTS.md`
- Тесты → `21_TESTING.md`
