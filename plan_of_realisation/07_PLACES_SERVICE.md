# 07 · places-service

> Управляет POI (точками интереса): публичными и пользовательскими, категориями, геопоиском, статусами модерации.

## 1. Bounded Context

places-service владеет: places, categories, place_categories, place_photos (id+order, файлы — в media-service).

Не владеет: модерационные решения и логика очереди (это `13_MODERATION_SERVICE.md`).

## 2. Доменные сущности

```go
type Place struct {
    ID          uuid.UUID
    Title       string                  // 1..120
    Description string                  // 0..2000
    Latitude    float64
    Longitude   float64
    Address     *string
    City        *string
    Country     *string
    AuthorID    *uuid.UUID              // null для seed-данных
    Status      PlaceStatus             // draft | pending_moderation | published | rejected | archived
    RejectReason *string
    Source      PlaceSource             // user | seed | imported
    CategoryIDs []uuid.UUID
    CoverMediaID *uuid.UUID
    PhotoMediaIDs []uuid.UUID           // упорядоченный список
    PublishedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Category struct {
    ID          uuid.UUID
    Slug        string  // unique, [a-z0-9-]
    Title       string
    Icon        string  // имя иконки в дизайне
    Color       string  // hex color из дизайна
    SortOrder   int
    Active      bool
}
```

## 3. Use cases

- `ListNearby(ctx, bbox, categoryIDs, q, limit) ([]Place, error)` — главная карта `MP-01`. Без cursor (только limit, max 200, всё в bbox).
- `Search(ctx, q, types, limit, cursor)` — текстовый поиск `MP-03`/`SR-02`. ILIKE по title + city. (Phase 3: ts_vector.)
- `GetPlace(ctx, id)` — для `MP-04`, проверяет статус (гость не видит non-published).
- `CreateDraft(ctx, userID, title, lat, lng, address, city, country)` — `MP-05`. Создаёт со статусом `draft`.
- `UpdateDraft(ctx, userID, placeID, ...)` — `MP-06`.
- `Submit(ctx, userID, placeID)` — переводит в `pending_moderation`, отправляет outbox `places.place_submitted`.
- `ListMine(ctx, userID, status, cursor, limit)` — для PF-будущего (мои предложения).
- Admin: `ListCategories`, `CreateCategory`, `UpdateCategory`, `DeactivateCategory` — `AD-04`.
- Internal: `ApprovePlace(ctx, placeID, adminID)`, `RejectPlace(ctx, placeID, adminID, reason)` — вызываются из moderation-service.

## 4. HTTP API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| GET | `/v1/places` | guest+ | `?bbox=lat1,lng1,lat2,lng2&category=slug&q=...&limit=200` |
| GET | `/v1/places/{id}` | guest+ (only published) или owner/admin | детали POI (MP-04) |
| POST | `/v1/places` | user | создать draft (MP-05) |
| PATCH | `/v1/places/{id}` | owner | обновить draft (MP-06) |
| POST | `/v1/places/{id}/submit` | owner | отправить на модерацию (MP-07 trigger) |
| GET | `/v1/places/categories` | guest+ | список активных категорий |
| POST | `/v1/places/categories` | admin | новая категория (AD-04) |
| PATCH | `/v1/places/categories/{id}` | admin | обновить |
| DELETE | `/v1/places/categories/{id}` | admin | деактивировать (soft-delete: active=false) |
| POST | `/v1/places/internal/{id}/approve` | internal | вызов из moderation-service |
| POST | `/v1/places/internal/{id}/reject` | internal | вызов из moderation-service |

## 5. БД-схема

```sql
SET search_path TO places, public;

CREATE TABLE places_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    icon TEXT NOT NULL,
    color TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE places_places (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    latitude FLOAT8 NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude FLOAT8 NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    address TEXT,
    city TEXT,
    country TEXT,
    author_id UUID,
    status TEXT NOT NULL CHECK (status IN ('draft','pending_moderation','published','rejected','archived')),
    reject_reason TEXT,
    source TEXT NOT NULL DEFAULT 'user' CHECK (source IN ('user','seed','imported')),
    cover_media_id UUID,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_places_status_pub ON places_places(status, published_at DESC) WHERE status='published';
CREATE INDEX ix_places_author ON places_places(author_id);
CREATE INDEX ix_places_geo ON places_places(latitude, longitude);
-- Для bbox-запросов: btree по обоим достаточно для MVP, потом мигрируем на gist+earthdistance.

CREATE TABLE places_place_categories (
    place_id UUID NOT NULL REFERENCES places_places(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES places_categories(id) ON DELETE RESTRICT,
    PRIMARY KEY (place_id, category_id)
);

CREATE TABLE places_place_photos (
    place_id UUID NOT NULL REFERENCES places_places(id) ON DELETE CASCADE,
    media_id UUID NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    PRIMARY KEY (place_id, media_id)
);

CREATE TABLE places_outbox (...);  -- стандартная outbox-таблица
```

### Seed-данные

В миграцию `20260502120000_seed_categories.sql` добавить базовые категории из дизайна (`MP-02` и описания):
- `museum` (Музей, teal)
- `gastro` (Гастро, coral)
- `nature` (Природа, sand)
- `architecture` (Архитектура, slate)
- `viewpoint` (Смотровая, outline)
- `park` (Парк, teal-300)

С seed мы получим непустую карту даже без admin-добавления категорий.

Дополнительно: 5-10 публичных POI как seed (Казанский Кремль, Эрмитаж и т.п. — из дизайн-мокапов).

## 6. Outbox

- `places.place_submitted` — payload: `{place_id, author_id, submitted_at}` → moderation-service (см. файл 13).
- `places.place_published` — payload: `{place_id, author_id, published_at}` → notifications-service (push автору + invalidate map cache).
- `places.place_rejected` — payload: `{place_id, author_id, reason}` → notifications-service.

## 7. Кэширование

- `GET /v1/places/categories` — кэш в Redis 1 час (`places:categories:v1`). Инвалидируется при `POST/PATCH/DELETE`.
- `GET /v1/places?bbox=...` — без кэша в MVP (тяжело инвалидировать на каждый publish).

## 8. Чек-лист

- [ ] `GET /v1/places?bbox=...` возвращает только `published` для гостя
- [ ] `POST /v1/places` создаёт `draft`, можно сабмитнуть только своё
- [ ] `POST /v1/places/{id}/submit` пишет в outbox, статус = `pending_moderation`
- [ ] Admin может управлять категориями (`AD-04`)
- [ ] Approve/Reject через `internal` API меняют статус и публикуют событие

## 9. Связанные файлы

- moderation-service → `13_MODERATION_SERVICE.md`
- media-service для фото POI → `10_MEDIA_SERVICE.md`
- БД → `15_DATABASE_SCHEMAS.md`
