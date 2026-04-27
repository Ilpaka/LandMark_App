# 14 · api-gateway

> Единая точка входа для клиента. Маршрутизирует `/v1/*` на нужный сервис, валидирует JWT, ставит rate-limit, добавляет внутренние заголовки.

## 1. Назначение

- Внешне: открыт на `:8080`, единственный endpoint наружу.
- Внутренне: знает Docker-сетевые имена сервисов и проксирует туда HTTP.
- Слой защиты: ни один внутренний сервис не должен быть доступен извне.

## 2. Структура каталогов

```
backend/api-gateway/
├── cmd/server/main.go
├── internal/
│   ├── config/config.go
│   ├── routing/                # таблица маршрутов
│   │   ├── routes.go
│   │   └── routes_test.go
│   ├── proxy/
│   │   └── reverse.go          # httputil.ReverseProxy с тонкой настройкой
│   └── middleware/
│       ├── jwt.go              # парсит Authorization, кэширует JWKS
│       ├── ratelimit.go        # Redis sliding-window
│       ├── cors.go
│       ├── requestid.go
│       └── logger.go
├── Dockerfile
├── docker-entrypoint.sh
├── Makefile
├── go.mod
└── go.sum
```

Без Postgres-зависимости. Подключается только к Redis (для rate-limit) и к auth-service (для `/v1/auth/jwks`).

## 3. Таблица маршрутизации

```go
// internal/routing/routes.go
type Route struct {
    Pattern   string  // "/v1/auth/*"
    Upstream  string  // "http://auth-service:8080"
    AuthMode  AuthMode
    RateLimit *RateLimit
}

type AuthMode int
const (
    AuthOptional   AuthMode = iota  // если есть токен — валидируем, нет — пропускаем
    AuthRequired                     // 401 если нет валидного токена
    AuthAdminOnly                    // 403 если role != admin
    AuthInternalOnly                 // ВСЕГДА 404 наружу — путь не должен быть proxied
)

var Routes = []Route{
    // auth (часть открыта)
    {"POST /v1/auth/register",       "http://auth-service:8080",   AuthOptional, RL{60, time.Hour}},
    {"POST /v1/auth/verify-email",   "http://auth-service:8080",   AuthOptional, RL{60, time.Hour}},
    {"POST /v1/auth/login",          "http://auth-service:8080",   AuthOptional, RL{20, time.Minute}},
    {"POST /v1/auth/refresh",        "http://auth-service:8080",   AuthOptional, RL{120, time.Hour}},
    {"POST /v1/auth/forgot-password","http://auth-service:8080",   AuthOptional, RL{10, time.Hour}},
    {"POST /v1/auth/reset-password", "http://auth-service:8080",   AuthOptional, RL{10, time.Hour}},
    {"POST /v1/auth/logout",         "http://auth-service:8080",   AuthRequired, nil},
    {"GET  /v1/auth/me",             "http://auth-service:8080",   AuthRequired, nil},
    {"GET  /v1/auth/sessions",       "http://auth-service:8080",   AuthRequired, nil},
    {"DELETE /v1/auth/sessions/*",   "http://auth-service:8080",   AuthRequired, nil},
    {"POST /v1/auth/me/email*",      "http://auth-service:8080",   AuthRequired, nil},
    {"POST /v1/auth/change-password","http://auth-service:8080",   AuthRequired, nil},
    {"DELETE /v1/auth/account",      "http://auth-service:8080",   AuthRequired, nil},
    {"GET  /v1/auth/jwks",           "http://auth-service:8080",   AuthOptional, nil},
    {"POST /v1/auth/phone/*",        "http://auth-service:8080",   AuthOptional, RL{20, time.Hour}},

    // admin
    {"GET  /v1/admin/*",             "http://auth-service:8080",   AuthAdminOnly, nil},
    {"POST /v1/admin/*",             "http://auth-service:8080",   AuthAdminOnly, nil},

    // profile
    {"GET    /v1/profile/me",        "http://profile-service:8080", AuthRequired, nil},
    {"PATCH  /v1/profile/me",        "http://profile-service:8080", AuthRequired, nil},
    {"GET    /v1/profile/me/privacy","http://profile-service:8080", AuthRequired, nil},
    {"PATCH  /v1/profile/me/privacy","http://profile-service:8080", AuthRequired, nil},
    {"GET    /v1/profile/by-nickname/*", "http://profile-service:8080", AuthRequired, nil},

    // places
    {"GET   /v1/places",             "http://places-service:8080",  AuthOptional, nil},
    {"GET   /v1/places/categories",  "http://places-service:8080",  AuthOptional, nil},
    {"GET   /v1/places/*",           "http://places-service:8080",  AuthOptional, nil},
    {"POST  /v1/places",             "http://places-service:8080",  AuthRequired, nil},
    {"PATCH /v1/places/*",           "http://places-service:8080",  AuthRequired, nil},
    {"POST  /v1/places/categories",  "http://places-service:8080",  AuthAdminOnly, nil},
    {"PATCH /v1/places/categories/*","http://places-service:8080",  AuthAdminOnly, nil},
    {"DELETE /v1/places/categories/*","http://places-service:8080", AuthAdminOnly, nil},

    // trips, journal, media, favorites, notifications, moderation, search
    // (полная таблица в коде; здесь сокращено)

    // internal: НИКОГДА не проксируем
    {"** /v1/*/internal/*",          "",                              AuthInternalOnly, nil},
}
```

> В реальности таблицу хранить в Go-структуре с явным mux'ом (gin или chi), не в строках.

## 4. JWT middleware

```go
// internal/middleware/jwt.go
type JWTMiddleware struct {
    JWKSURL    string
    Cache      *jwksCache  // TTL 1h, refresh on 401
    Issuer     string
    Audience   string
}

func (m *JWTMiddleware) Verify(c *gin.Context) {
    h := c.GetHeader("Authorization")
    if !strings.HasPrefix(h, "Bearer ") { ...optional or 401 }
    tok := h[7:]
    claims, err := m.verify(tok)
    if err != nil { c.AbortWithStatusJSON(401, ErrorBody{"error":"unauthorized"}); return }
    // blacklist check (Redis)
    if blacklisted, _ := m.Redis.Has(ctx, claims.JTI); blacklisted {
        c.AbortWithStatusJSON(401, ...); return
    }
    // ставим внутренние заголовки для downstream
    c.Request.Header.Set("X-User-Id", claims.Sub.String())
    c.Request.Header.Set("X-Session-Id", claims.Session.String())
    c.Request.Header.Set("X-Role", string(claims.Role))
    // удаляем Authorization чтобы не утекало внутрь
    c.Request.Header.Del("Authorization")
    c.Next()
}
```

## 5. Rate-limit middleware

Sliding-window через Redis:
- Ключ: `rl:{route_id}:{user_id_or_ip}`
- Реализация — как в Tennis `adapters/redis/client.go` (унаследовано).
- При превышении: 429 + заголовки `X-RateLimit-{Limit,Remaining,Reset}`.
- `fail-closed`: если Redis недоступен — отвергаем с 503.

## 6. CORS

В dev: `Access-Control-Allow-Origin: *`. В prod: whitelist из env `CORS_ORIGINS=https://...`.

## 7. Request-ID

- Если клиент прислал `X-Request-Id` — используем.
- Иначе генерируем UUID и добавляем в request.
- Логируем + пробрасываем в downstream.

## 8. Конфигурация

| Env | Default | Назначение |
|---|---|---|
| HTTP_ADDR | `:8080` | публичный порт |
| AUTH_JWKS_URL | `http://auth-service:8080/v1/auth/jwks` | для JWT |
| JWT_ISSUER | `landmark.app` | проверка iss |
| JWT_AUDIENCE | `landmark-clients` | проверка aud |
| REDIS_URL | `redis://redis:6379/0` | для rate-limit + blacklist |
| CORS_ORIGINS | `*` | список разрешённых |
| OTEL_SERVICE_NAME | `landmark-api-gateway` | |
| LOG_LEVEL | `info` | |

## 9. Ошибки

Gateway возвращает свои ошибки только для:
- 401 unauthorized — нет/невалидный токен
- 403 forbidden — нужен admin
- 404 not_found — нет такого маршрута
- 429 rate_limited — превышен лимит
- 502 bad_gateway — upstream вернул 5xx или недоступен
- 504 gateway_timeout — таймаут к upstream (default: 30s, upload-роуты: 120s)

Все остальные ошибки приходят из upstream-сервисов «как есть» (с тем же status code и body).

## 10. Деплой

В docker-compose это единственный сервис, у которого `ports: ["8080:8080"]`. Все остальные `expose: ["8080"]` (только для внутренней сети).

## 11. Связанные файлы

- Auth → `05_AUTH_SERVICE.md`
- Все downstream-сервисы → `06..13`
- Docker → `16_DOCKER_INFRA.md`
