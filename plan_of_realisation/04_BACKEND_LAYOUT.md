# 04 · Структура и принципы каждого Go-сервиса

> Все сервисы используют **одну и ту же структуру каталогов**. Это позволяет ИИ генерировать новые сервисы по шаблону. Шаблон взят из `Tennis_app/backend/auth-service` (см. `05_AUTH_SERVICE.md`).

## 1. Принципы Clean Architecture в Go

```
┌─────────────────────────────────────────────────┐
│                  cmd/                           │  ← entry points (main + migrate)
├─────────────────────────────────────────────────┤
│  internal/httpserver/   internal/config/        │  ← infra glue
├─────────────────────────────────────────────────┤
│              internal/modules/<domain>/         │
│  ┌───────────────────────────────────────────┐  │
│  │ domain/    — entities, value objects, errs│  │  inner ring
│  │ ports/     — interfaces (Store, Sender..) │  │
│  │ app/       — use cases (Service)          │  │  use cases
│  │ adapters/  — postgres, redis, jwt, sms,...│  │  outer ring
│  │ http/      — handlers, middleware, router │  │
│  │ worker/    — outbox poller, cron tasks    │  │
│  │ metrics/   — Prometheus metrics           │  │
│  │ audit/     — auditing helpers             │  │
│  │ crypto/    — service-local crypto helpers │  │
│  │ mocks/     — generated mocks (gomock)     │  │
│  └───────────────────────────────────────────┘  │
├─────────────────────────────────────────────────┤
│         api/openapi/<service>.yaml              │  ← contract
├─────────────────────────────────────────────────┤
│         db/migrations/*.sql                     │  ← schema
├─────────────────────────────────────────────────┤
│         tests/integration/*  tests/e2e/*        │  ← black-box tests
└─────────────────────────────────────────────────┘
```

**Правила зависимостей** (никогда не нарушать):
- `domain` ничего не импортирует, кроме stdlib + `github.com/google/uuid`.
- `ports` импортируют только `domain`.
- `app` импортируют `ports` + `domain`. **Не импортируют адаптеры напрямую.**
- `adapters` имплементируют интерфейсы из `ports`.
- `http` импортирует `app` (для вызова use-cases) + `domain` (для типов ошибок).
- `worker` импортирует `app` или `ports` напрямую (если worker — это просто потребитель Store).
- `cmd/server/main.go` собирает граф зависимостей: создаёт адаптеры, прокидывает в Service, передаёт в http.Mount.

## 2. Скелет каталогов сервиса

```
backend/<service-name>/
├── api/
│   └── openapi/
│       └── <service>.yaml
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── migrate/
│       └── main.go
├── db/
│   └── migrations/
│       ├── 20260501120000_init_schema.sql
│       └── ...
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── httpserver/
│   │   ├── server.go         # MountX(...) -> *Stack
│   │   ├── middleware.go     # request-id, logging, recovery, otel
│   │   └── swagger.go        # /docs если cfg.DevSwagger
│   └── modules/
│       └── <domain>/
│           ├── domain/
│           ├── ports/
│           │   ├── ports.go
│           │   └── generate.go   # //go:generate mockgen ...
│           ├── app/
│           ├── adapters/
│           │   ├── postgres/
│           │   ├── redis/
│           │   ├── jwt/         # только для auth
│           │   ├── sms/         # только для auth/notifications
│           │   ├── email/
│           │   └── <other>/
│           ├── http/
│           ├── worker/
│           ├── metrics/
│           ├── audit/           # для сервисов с auditing
│           ├── crypto/          # для сервисов с crypto-операциями
│           └── mocks/
├── tests/
│   ├── e2e/
│   ├── integration/
│   ├── security/
│   └── testutil/
├── scripts/
│   └── schemathesis.sh    # contract-fuzz через openapi
├── docker-compose.yml     # сервис-локальный (опц., чаще пользуемся корневым)
├── docker-entrypoint.sh
├── Dockerfile
├── go.mod
├── go.sum
└── Makefile
```

## 3. Шаблонные файлы

### 3.1 Makefile (общий для всех сервисов)

```makefile
.PHONY: tidy keys test test-cover cover-check test-integration test-e2e lint sec migrate-up migrate-down run

SERVICE_NAME ?= $(shell basename $(CURDIR))

tidy:
	go mod tidy

keys:  # only for auth-service
	mkdir -p keys && openssl genrsa -out keys/jwt_rsa.pem 2048 && openssl rsa -in keys/jwt_rsa.pem -pubout -out keys/jwt_rsa.pub.pem

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

test:
	go test ./... -count=1 -short

test-cover:
	go test ./... -count=1 -short -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 5

test-integration:
	go test ./tests/integration/... -count=1 -tags=integration -timeout=400s

test-e2e:
	go test ./tests/e2e/... -count=1 -tags=e2e -timeout=600s -parallel 1

lint:
	go vet ./...

sec:
	go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet -exclude-generated -severity=high ./...

run:
	go run ./cmd/server
```

### 3.2 Dockerfile (общий шаблон)

```dockerfile
# Build from repository root: docker compose build
FROM golang:1.25-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src/backend
COPY backend/<SERVICE>/go.mod backend/<SERVICE>/go.sum ./<SERVICE>/
COPY backend/pkg ./pkg/
WORKDIR /src/backend/<SERVICE>
RUN go mod download
COPY backend/<SERVICE>/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=build /out/server /out/migrate /app/
COPY backend/<SERVICE>/db/migrations /app/db/migrations
COPY backend/<SERVICE>/api/openapi /app/api/openapi
COPY backend/<SERVICE>/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh && chown -R app:app /app
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080 \
    GOOSE_MIGRATION_DIR=/app/db/migrations
ENTRYPOINT ["/app/docker-entrypoint.sh"]
```

> Для auth-service дополнительно: `mkdir -p /app/keys && chown app:app /app/keys`, плюс генерация ключей в entrypoint.

### 3.3 docker-entrypoint.sh (общий)

```sh
#!/bin/sh
set -e
echo "Running database migrations..."
/app/migrate up
echo "Starting server..."
exec /app/server
```

## 4. Общие пакеты `backend/pkg/`

### 4.1 `backend/pkg/observability` (унаследован из Tennis)

- `SetupOTel(ctx, serviceName, otlpEndpoint) (shutdown func(ctx) error, err error)` — поднимает tracer provider; export через otlphttp.
- Используем otelgin для автоматического создания спанов на каждый HTTP-запрос.

> **Действие:** скопировать **как есть** из `Tennis_app/backend/pkg/observability` в `LandMark_App/backend/pkg/observability`. Поправить `module github.com/ilpaka/tennis_app/...` → `module github.com/ilpaka/landmark_app/backend/pkg/observability`.

### 4.2 `backend/pkg/httpx` (новый)

Утилиты, которые понадобятся всем сервисам, но в Tennis они дублируются. Вынесем сюда:

- `WriteError(c *gin.Context, err error)` — единая запись `ErrorBody` (см. `17_API_CONTRACTS.md`).
- `RequestID(c *gin.Context) string` — извлекает или генерирует `X-Request-Id`.
- `Pagination(c *gin.Context) (cursor string, limit int)` — валидирует `?cursor=&limit=`.
- `BindOrFail(c, &req) bool` — обёртка над ShouldBindJSON.

### 4.3 `backend/pkg/authmid` (новый)

JWT middleware для всех сервисов кроме auth: парсит `Authorization`, верифицирует через **JWKS endpoint auth-service**, кэширует ключи в памяти на 1 час, ставит claims в gin.Context. Используется и в Gateway, и в каждом downstream-сервисе (defense-in-depth: если кто-то получит прямой доступ — всё равно проверка).

### 4.4 `backend/pkg/eventbus` (новый, для outbox + Redis Pub/Sub)

```go
type Event struct {
    AggregateType string
    EventType     string
    AggregateID   string
    Payload       map[string]any
    OccurredAt    time.Time
}
type Publisher interface { Publish(ctx context.Context, ev Event) error }
type Subscriber interface { Subscribe(ctx context.Context, pattern string, handler func(Event) error) error }
```

Реализация — Redis Pub/Sub. Outbox-poller каждого сервиса перекладывает события из Postgres-таблицы `*_outbox` в Redis-канал `events.<aggregate_type>.<event_type>`.

## 5. Конвенции кода

### 5.1 Именование

- Модули в `internal/modules/` — единственное число имени домена: `auth`, `place`, `trip`, `journal`, `media`, `favorite`, `notification`, `moderation`.
- HTTP-пакеты: `<domain>http` (например, `placehttp`), чтобы избежать коллизий с stdlib `net/http`.
- Mock-файлы: `mocks/mock_<interface>.go` через `mockgen`.

### 5.2 Ошибки

- Все доменные ошибки — типизированные значения в `domain/errors.go`:
  ```go
  var (
      ErrNotFound     = errors.New("not found")
      ErrConflict     = errors.New("conflict")
      ErrBadRequest   = errors.New("bad request")
      ErrUnauthorized = errors.New("unauthorized")
      ErrForbidden    = errors.New("forbidden")
      ErrRateLimited  = errors.New("rate limited")
      ErrUnavailable  = errors.New("unavailable")
  )
  ```
- HTTP-слой переводит их в HTTP-коды через единую функцию `httpx.WriteError`.

### 5.3 Логи

- `slog.JSONHandler` на stdout.
- Уровень из env `LOG_LEVEL` (`debug|info|warn|error`).
- В каждый запрос добавляется `request_id` через middleware.

### 5.4 Контекст

- `context.Context` как первый аргумент **во все методы пользовательских слоёв (app, ports)**. В адаптерах — куда передаётся стандартно.
- В контекст кладётся `request_id` (через middleware) и `tx` (через `WithTx`).

### 5.5 Транзакции

- Pattern из Tennis: `Store.WithTx(ctx, func(ctx, tx pgx.Tx) error { ... })`.
- В тестах: `dockertest` поднимает Postgres и Redis (см. `21_TESTING.md`).

## 6. Skeleton: `cmd/server/main.go` (универсальный)

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/ilpaka/landmark_app/backend/<service>/internal/config"
    "github.com/ilpaka/landmark_app/backend/<service>/internal/httpserver"
    redisx "github.com/ilpaka/landmark_app/backend/<service>/internal/modules/<domain>/adapters/redis"
    "github.com/ilpaka/landmark_app/backend/pkg/observability"
)

func main() {
    log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
    slog.SetDefault(log)

    cfg, err := config.Load()
    if err != nil { log.Error("config", "err", err); os.Exit(1) }
    if cfg.Env == "production" { gin.SetMode(gin.ReleaseMode) }

    ctx := context.Background()
    shutdownOTel, err := observability.SetupOTel(ctx, cfg.OTELServiceName, cfg.OTELExporterOTLPEndpoint)
    if err != nil { log.Error("otel", "err", err); os.Exit(1) }
    defer func() {
        c, cancel := context.WithTimeout(context.Background(), 5*time.Second); defer cancel()
        if err := shutdownOTel(c); err != nil { log.Warn("otel shutdown", "err", err) }
    }()

    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil { log.Error("pg", "err", err); os.Exit(1) }
    defer pool.Close()

    rdb, err := redisx.New(cfg.RedisURL)
    if err != nil { log.Error("redis", "err", err); os.Exit(1) }

    stack, err := httpserver.Mount<Domain>(cfg, log, pool, rdb)
    if err != nil { log.Error("mount", "err", err); os.Exit(1) }

    srvCtx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go stack.Worker.Run(srvCtx)  // outbox poller, etc.

    go func() {
        if err := stack.Engine.Run(cfg.HTTPAddr); err != nil { log.Error("http", "err", err); os.Exit(1) }
    }()

    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
    cancel()
}
```

## 7. Универсальный config.go

```go
type Config struct {
    Env                      string  // APP_ENV
    HTTPAddr                 string  // HTTP_ADDR
    DatabaseURL              string
    RedisURL                 string
    JWTPublicKeyPath         string  // для проверки токенов от auth (downstream сервисы)
    OTELServiceName          string
    OTELExporterOTLPEndpoint string
    DevSwagger               bool
    OpenAPIYAML              string  // путь до openapi
    // ... сервис-специфичные поля
}
```

Никогда не хардкодить URL и таймауты — всё из env с дефолтами.

## 8. Связанные файлы

- Реализация конкретных сервисов → `05..14`
- Полные миграции БД → `15_DATABASE_SCHEMAS.md`
- Docker-инфра → `16_DOCKER_INFRA.md`
