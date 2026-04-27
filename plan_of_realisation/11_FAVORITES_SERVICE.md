# 11 · favorites-service

> Избранное: пользователь сохраняет POI/Trip/Entry в личный список и/или в коллекции.

## 1. Bounded Context

favorites-service владеет: favorites (user_id, target_type, target_id), collections, collection_items.

## 2. Сущности

```go
type FavoriteTargetType string  // place | trip | entry
type Favorite struct {
    UserID     uuid.UUID
    TargetType FavoriteTargetType
    TargetID   uuid.UUID
    CollectionID *uuid.UUID  // null = «без коллекции»
    Note       *string
    CreatedAt  time.Time
}

type Collection struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Title     string
    Color     string  // hex
    IsDefault bool
    CreatedAt time.Time
}
```

## 3. Use cases

- `Add(ctx, userID, targetType, targetID, collectionID?, note?)` — FV-01 add. Идемпотентно (дубликат → 200 OK).
- `Remove(ctx, userID, targetType, targetID)`.
- `List(ctx, userID, targetType?, collectionID?, cursor, limit)` — FV-01.
- `ListIDs(ctx, userID, targetType, ids)` — bulk-проверка «лайкнуто ли это» для batch UI.
- `CreateCollection`, `UpdateCollection`, `DeleteCollection` (default нельзя удалить).
- `MoveTo(ctx, userID, targetType, targetID, newCollectionID)`.

## 4. HTTP API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| GET | `/v1/favorites` | user | `?type=&collection=&cursor=&limit=` (FV-01) |
| POST | `/v1/favorites` | user | `{target_type, target_id, collection_id?, note?}` |
| DELETE | `/v1/favorites/{type}/{target_id}` | user | удалить |
| POST | `/v1/favorites/exists` | user | bulk: `{type, ids: []}` → `{[id]: bool}` |
| GET | `/v1/favorites/collections` | user | список |
| POST | `/v1/favorites/collections` | user | создать |
| PATCH | `/v1/favorites/collections/{id}` | user | переименовать/перекрасить |
| DELETE | `/v1/favorites/collections/{id}` | user | удалить (предметы → default) |
| POST | `/v1/favorites/move` | user | `{target_type, target_id, collection_id}` |

## 5. БД-схема

```sql
SET search_path TO favorites, public;

CREATE TABLE favorites_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#0E7C7B',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_collection_default_per_user ON favorites_collections(user_id) WHERE is_default;

CREATE TABLE favorites_favorites (
    user_id UUID NOT NULL,
    target_type TEXT NOT NULL CHECK (target_type IN ('place','trip','entry')),
    target_id UUID NOT NULL,
    collection_id UUID,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, target_type, target_id)
);
CREATE INDEX ix_fav_user ON favorites_favorites(user_id, created_at DESC);
CREATE INDEX ix_fav_collection ON favorites_favorites(collection_id);
```

При первом обращении пользователя автоматически создаётся default-collection «Сохранённое» (триггер на `INSERT favorites_favorites` или ленивая инициализация в use-case).

## 6. Чек-лист

- [ ] Add идемпотентен
- [ ] Removing nonexistent → 204 No Content (не 404)
- [ ] Default collection нельзя удалить (409)
- [ ] List поддерживает фильтрацию по type и collection

## 7. Связанные файлы

- places (детали POI для FV) → клиент делает 2 запроса (favorites + places batch)
- journal (entry detail для FV) — то же
- БД → `15`
