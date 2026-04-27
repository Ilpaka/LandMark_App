# 09 · journal-service

> Записи журнала: текстовые посты с фотографиями, тегами и привязкой к поездке/точке.

## 1. Bounded Context

journal-service владеет: entries, entry_tags, entry_media (id+order, файлы — в media).

Не владеет: trips, places, media-файлы. Ссылается через `trip_id`, `place_id`, `media_ids`.

## 2. Сущности

```go
type Entry struct {
    ID          uuid.UUID
    AuthorID    uuid.UUID
    TripID      *uuid.UUID  // может быть запись вне поездки
    PlaceID     *uuid.UUID  // ссылка на POI, опц.
    Latitude    *float64    // если без place_id, координаты явно
    Longitude   *float64
    Title       *string
    Body        string      // markdown, до 10000 chars
    Mood        *Mood       // happy | excited | calm | tired | sad
    Rating      *int        // 1..5, опционально
    OccurredAt  time.Time   // когда событие произошло (≠ created_at)
    MediaIDs    []uuid.UUID // упорядоченные
    Tags        []string    // ["отлично", "природа"]
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}
```

## 3. Use cases

- `Create(ctx, authorID, input)` — JR-03. Body валидируется. media_ids проверяются на ownership через media-internal API.
- `Update(ctx, authorID, entryID, input)` — JR-02 edit.
- `Delete(ctx, authorID, entryID)` — soft (deleted_at).
- `ListFeed(ctx, authorID, tripID, cursor, limit)` — JR-01. Если `tripID` задан, фильтр.
- `ListByTrip(ctx, viewerID, tripID, cursor, limit)` — для TR-04.
- `Get(ctx, viewerID, entryID)` — JR-02.
- `AddMedia(ctx, authorID, entryID, mediaID, order)` — добавление фото после создания.

## 4. HTTP API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| GET | `/v1/journal/entries` | user | `?trip_id=&cursor=&limit=` (JR-01) |
| POST | `/v1/journal/entries` | user | создать (JR-03) |
| GET | `/v1/journal/entries/{id}` | author | (JR-02) |
| PATCH | `/v1/journal/entries/{id}` | author | редактировать |
| DELETE | `/v1/journal/entries/{id}` | author | удалить |
| GET | `/v1/journal/entries/{id}/media` | author | (JR-04) |
| POST | `/v1/journal/entries/{id}/media` | author | прикрепить медиа: `{media_id, sort_order}` |
| DELETE | `/v1/journal/entries/{id}/media/{mid}` | author | открепить |
| GET | `/v1/journal/internal/by-trip/{tripID}/counts` | internal | счётчики (entries, photos) |

## 5. БД-схема

```sql
SET search_path TO journal, public;

CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL,
    trip_id UUID,
    place_id UUID,
    latitude FLOAT8,
    longitude FLOAT8,
    title TEXT,
    body TEXT NOT NULL DEFAULT '',
    mood TEXT CHECK (mood IS NULL OR mood IN ('happy','excited','calm','tired','sad')),
    rating SMALLINT CHECK (rating IS NULL OR rating BETWEEN 1 AND 5),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX ix_entries_author ON journal_entries(author_id, occurred_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX ix_entries_trip ON journal_entries(trip_id, occurred_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE journal_entry_media (
    entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    media_id UUID NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    PRIMARY KEY (entry_id, media_id)
);

CREATE TABLE journal_entry_tags (
    entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (entry_id, tag)
);

CREATE TABLE journal_outbox (...);
```

## 6. Outbox события

- `journal.entry_created` — `{author_id, entry_id, trip_id, photos_count, occurred_at}` → trips-service инкрементит счётчики; notifications.
- `journal.entry_deleted` — для счётчиков.

## 7. Чек-лист

- [ ] Create + List + Get + Delete покрыты unit + integration тестами
- [ ] media_ids валидируются через media-internal перед сохранением
- [ ] Привязка к trip ограничена ownership: проверка через trips-internal
- [ ] Rate-limit: 60 entries/час на user (Redis sliding window)

## 8. Связанные файлы

- trips → `08_TRIPS_SERVICE.md`
- media → `10_MEDIA_SERVICE.md`
- БД → `15_DATABASE_SCHEMAS.md`
