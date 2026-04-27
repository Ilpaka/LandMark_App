# 17 · API контракты (OpenAPI 3.0)

> Каждый сервис содержит `api/openapi/<service>.yaml`. Этот файл задаёт **общие правила** и приводит шаблон.

## 1. Общие модели

Эти модели включаются в каждый OpenAPI-файл (через `$ref` или копирование):

```yaml
components:
  schemas:
    ErrorBody:
      type: object
      required: [error]
      properties:
        error: { type: string, description: "Машинно-читаемый код ошибки" }
        code:  { type: string, description: "Опциональный технический код" }
        details:
          type: object
          additionalProperties: true
          description: "Доп. контекст для отладки"

    StatusOK:
      type: object
      required: [status]
      properties:
        status: { type: string, example: ok }

    PageInfo:
      type: object
      required: [next_cursor, has_more]
      properties:
        next_cursor: { type: string, nullable: true }
        has_more:    { type: boolean }

  responses:
    BadRequest:        { description: Bad Request,        content: { application/json: { schema: { $ref: "#/components/schemas/ErrorBody" } } } }
    Unauthorized:      { description: Unauthorized,       content: { application/json: { schema: { $ref: "#/components/schemas/ErrorBody" } } } }
    Forbidden:         { description: Forbidden,          content: { application/json: { schema: { $ref: "#/components/schemas/ErrorBody" } } } }
    NotFound:          { description: Not Found,          content: { application/json: { schema: { $ref: "#/components/schemas/ErrorBody" } } } }
    Conflict:          { description: Conflict,           content: { application/json: { schema: { $ref: "#/components/schemas/ErrorBody" } } } }
    RateLimited:       { description: Too Many Requests,  content: { application/json: { schema: { $ref: "#/components/schemas/ErrorBody" } } } }
    ServiceUnavailable:{ description: Service Unavailable,content: { application/json: { schema: { $ref: "#/components/schemas/ErrorBody" } } } }
```

## 2. Конвенции ошибок (значения поля `error`)

| Код | HTTP | Когда |
|---|---|---|
| `validation_error` | 400 | Невалидный JSON, поле, формат |
| `unauthorized` | 401 | Нет/истёкший/невалидный токен |
| `forbidden` | 403 | Нет прав (роль/ownership) |
| `not_found` | 404 | Ресурс не найден |
| `conflict` | 409 | Конфликт (дубликат, состояние) |
| `rate_limited` | 429 | Лимит превышен |
| `unavailable` | 503 | Зависимость недоступна |
| `internal` | 500 | Неожиданное (не мапятся доменные ошибки) |

## 3. Аутентификация

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

Эндпоинты с `AuthOptional` или public — добавляют `security: []` в свой operation.

## 4. Шаблон спецификации сервиса

```yaml
openapi: 3.0.3
info:
  title: LandMark <Service> API
  version: 0.1.0
  description: |
    Внутренний REST API сервиса <Service>. Все эндпоинты префиксуются /v1/<service>/.
servers:
  - url: /
paths:
  /healthz:
    get:
      summary: Liveness/readiness
      operationId: healthz
      security: []
      responses:
        '200': { description: OK }
        '503': { description: Зависимость недоступна, content: { application/json: { schema: { $ref: '#/components/schemas/ErrorBody' } } } }

  # ... endpoints
```

## 5. Эндпоинты по сервисам (минимальный список — каждый расширяется в своём YAML)

### Auth (`api/openapi/auth.yaml`)

См. `05_AUTH_SERVICE.md` §2.11. Унаследовано из Tennis + новый `/v1/auth/register`.

### Profile (`api/openapi/profile.yaml`)

```
GET    /v1/profile/me
PATCH  /v1/profile/me
GET    /v1/profile/by-nickname/{nick}
GET    /v1/profile/me/privacy
PATCH  /v1/profile/me/privacy
POST   /v1/profile/internal/nicknames/reserve
POST   /v1/profile/internal/nicknames/{rid}/confirm
DELETE /v1/profile/internal/nicknames/{rid}
GET    /v1/profile/internal/{user_id}
```

Модель `Profile`:
```yaml
Profile:
  type: object
  required: [user_id, nickname, display_name]
  properties:
    user_id: { type: string, format: uuid }
    nickname: { type: string, minLength: 3, maxLength: 32, pattern: '^[a-z0-9_]+$' }
    display_name: { type: string, minLength: 1, maxLength: 64 }
    bio: { type: string, nullable: true, maxLength: 500 }
    avatar_media_id: { type: string, format: uuid, nullable: true }
    city: { type: string, nullable: true }
    country: { type: string, nullable: true }
```

### Places (`api/openapi/places.yaml`)

```
GET    /v1/places                         # ?bbox=&category=&q=&limit=
GET    /v1/places/{id}
POST   /v1/places                         # draft
PATCH  /v1/places/{id}
POST   /v1/places/{id}/submit
GET    /v1/places/categories
POST   /v1/places/categories              # admin
PATCH  /v1/places/categories/{id}         # admin
DELETE /v1/places/categories/{id}         # admin
POST   /v1/places/internal/{id}/approve
POST   /v1/places/internal/{id}/reject
```

### Trips (`api/openapi/trips.yaml`)

```
GET    /v1/trips                          # ?status=&cursor=&limit=
POST   /v1/trips
GET    /v1/trips/{id}
PATCH  /v1/trips/{id}
DELETE /v1/trips/{id}
GET    /v1/trips/{id}/stops
POST   /v1/trips/{id}/stops
PATCH  /v1/trips/{id}/stops/{sid}
DELETE /v1/trips/{id}/stops/{sid}
POST   /v1/trips/{id}/stops/{sid}/visit
GET    /v1/trips/{id}/routes
POST   /v1/trips/{id}/routes
PATCH  /v1/trips/{id}/routes/{rid}
DELETE /v1/trips/{id}/routes/{rid}
POST   /v1/trips/{id}/routes/{rid}/stops
GET    /v1/trips/{id}/routes/{rid}/stops
POST   /v1/trips/internal/{id}/counters
```

### Journal (`api/openapi/journal.yaml`)

```
GET    /v1/journal/entries                # ?trip_id=&cursor=&limit=
POST   /v1/journal/entries
GET    /v1/journal/entries/{id}
PATCH  /v1/journal/entries/{id}
DELETE /v1/journal/entries/{id}
GET    /v1/journal/entries/{id}/media
POST   /v1/journal/entries/{id}/media
DELETE /v1/journal/entries/{id}/media/{mid}
GET    /v1/journal/internal/by-trip/{tripID}/counts
```

### Media (`api/openapi/media.yaml`)

```
POST   /v1/media/uploads
POST   /v1/media/uploads/{id}/finalize
GET    /v1/media/{id}
DELETE /v1/media/{id}
GET    /v1/media/internal/{id}/owner
GET    /v1/media/internal/by-ids
```

### Favorites (`api/openapi/favorites.yaml`)

```
GET    /v1/favorites                      # ?type=&collection=&cursor=&limit=
POST   /v1/favorites
DELETE /v1/favorites/{type}/{target_id}
POST   /v1/favorites/exists
GET    /v1/favorites/collections
POST   /v1/favorites/collections
PATCH  /v1/favorites/collections/{id}
DELETE /v1/favorites/collections/{id}
POST   /v1/favorites/move
```

### Notifications (`api/openapi/notifications.yaml`)

```
GET    /v1/notifications
POST   /v1/notifications/read
POST   /v1/notifications/devices
DELETE /v1/notifications/devices/{id}
GET    /v1/notifications/preferences
PATCH  /v1/notifications/preferences
POST   /v1/notifications/internal/enqueue
```

### Moderation (`api/openapi/moderation.yaml`)

```
GET    /v1/moderation/queue
GET    /v1/moderation/items/{id}
POST   /v1/moderation/items/{id}/assign
POST   /v1/moderation/items/{id}/decision
GET    /v1/moderation/stats
```

### Search (вынесен в places-service для MVP)

```
GET    /v1/search                         # ?q=&types=place,trip,entry&cursor=&limit=
```

В Phase 2 — отдельный search-service с Postgres FTS.

## 6. Контракт-фьюзинг

Каждый сервис содержит `scripts/schemathesis.sh` (унаследовано из Tennis):

```sh
#!/bin/sh
schemathesis run --base-url=http://localhost:8080 \
    --hypothesis-deadline=2000 \
    --checks=all \
    api/openapi/<service>.yaml
```

В CI (опц.) запускается `make schemathesis` после поднятия сервиса.

## 7. Связанные файлы

- Каждый сервис → `05..13`
- Gateway, который защищает routes → `14_API_GATEWAY.md`
- Тестирование → `21_TESTING.md`
