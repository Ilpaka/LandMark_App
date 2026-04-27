# 08 · trips-service

> Поездки пользователя с датами, регионом, точками маршрута и обложками. Записи журнала живут в отдельном journal-service, но привязаны через `trip_id`.

## 1. Bounded Context

trips-service владеет: trips, trip_stops, trip_routes, trip_route_stops, trip_collaborators (для будущего).

Не владеет: записи журнала (см. `09`), фото поездки (`10`), POI (`07`). Поездка может ссылаться на POI через `trip_stops.place_id`.

## 2. Доменные сущности

```go
type Trip struct {
    ID         uuid.UUID
    OwnerID    uuid.UUID
    Title      string             // 1..120
    Subtitle   *string            // короткое описание/регион "Майские в Карелии"
    StartDate  *time.Time         // дата начала (только дата, не datetime)
    EndDate    *time.Time
    CoverMediaID *uuid.UUID
    Status     TripStatus         // draft | planned | in_progress | completed | archived
    Region     *string            // human-readable e.g. "Карелия"
    CenterLat  *float64           // для карты-обложки
    CenterLng  *float64
    StopsCount int                // denormalized
    EntriesCount int              // denormalized (синхронизируется через events от journal)
    PhotosCount int               // denormalized
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type TripStop struct {
    ID         uuid.UUID
    TripID     uuid.UUID
    PlaceID    *uuid.UUID  // null = свободная точка (не из POI)
    Title      string
    Latitude   float64
    Longitude  float64
    PlannedAt  *time.Time  // когда планируем посетить (только дата)
    VisitedAt  *time.Time  // фактически посещено (set on visit)
    Notes      *string
    Order      int
    CreatedAt  time.Time
}

type Route struct {
    ID         uuid.UUID
    TripID     uuid.UUID
    Title      string             // "День 1: центр города"
    Day        *int               // номер дня от старта поездки, опц.
    Color      string             // hex
    CreatedAt  time.Time
}

type RouteStop struct {
    RouteID    uuid.UUID
    StopID     uuid.UUID
    Order      int
    PRIMARY KEY (RouteID, StopID)
}
```

`TripStatus`:
- `draft` — пользователь начал создание (TR-02), но не закончил
- `planned` — даты установлены, поездка в будущем
- `in_progress` — текущее время попадает между start/end
- `completed` — после endDate
- `archived` — пользователь скрыл

Статус обновляется фоновой задачей раз в час (см. `worker/`).

## 3. Use cases

- `CreateDraft(ctx, ownerID, title)` → возвращает `Trip` со статусом `draft`.
- `UpdateMeta(ctx, ownerID, tripID, ...)` — обновить любые поля кроме status/owner.
- `SetDates(ctx, ownerID, tripID, start, end)` — TR-03; авто-переход `draft→planned`.
- `AddStop(ctx, ownerID, tripID, stop)` — для построения маршрута.
- `RemoveStop(ctx, ownerID, tripID, stopID)`.
- `MarkVisited(ctx, ownerID, stopID, at)` — фиксация посещения.
- `CreateRoute(ctx, ownerID, tripID, title, day, color)` — RT-02.
- `AddStopsToRoute(ctx, ownerID, routeID, stopIDs)`.
- `ListMine(ctx, ownerID, status, cursor, limit)` — TR-01.
- `Get(ctx, viewerID, tripID)` — TR-04.
- `Archive(ctx, ownerID, tripID)`.

Internal:
- `IncrementEntriesCount(ctx, tripID, delta)` — вызывается из journal-service event handler.
- `IncrementPhotosCount(ctx, tripID, delta)`.

## 4. HTTP API

| Метод | Путь | Auth | Назначение |
|---|---|---|---|
| GET | `/v1/trips` | user | `?status=&cursor=&limit=` (TR-01) |
| POST | `/v1/trips` | user | создать draft (TR-02) |
| GET | `/v1/trips/{id}` | user (owner) | детали (TR-04) |
| PATCH | `/v1/trips/{id}` | owner | обновить (TR-03 — даты) |
| DELETE | `/v1/trips/{id}` | owner | удалить (soft, status=archived) |
| GET | `/v1/trips/{id}/stops` | owner | список точек (TR-05) |
| POST | `/v1/trips/{id}/stops` | owner | добавить точку |
| PATCH | `/v1/trips/{id}/stops/{sid}` | owner | редактировать |
| DELETE | `/v1/trips/{id}/stops/{sid}` | owner | удалить |
| POST | `/v1/trips/{id}/stops/{sid}/visit` | owner | пометить посещённой |
| GET | `/v1/trips/{id}/routes` | owner | список маршрутов (RT-01) |
| POST | `/v1/trips/{id}/routes` | owner | создать маршрут (RT-02) |
| PATCH | `/v1/trips/{id}/routes/{rid}` | owner | обновить |
| DELETE | `/v1/trips/{id}/routes/{rid}` | owner | удалить |
| POST | `/v1/trips/{id}/routes/{rid}/stops` | owner | привязать stops (body: `{stop_ids: []}`) |
| GET | `/v1/trips/{id}/routes/{rid}/stops` | owner | упорядоченные точки (RT-04) |
| POST | `/v1/trips/internal/{id}/counters` | internal | инкремент счётчиков |

## 5. БД-схема

```sql
SET search_path TO trips, public;

CREATE TABLE trips_trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    title TEXT NOT NULL,
    subtitle TEXT,
    start_date DATE,
    end_date DATE,
    cover_media_id UUID,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','planned','in_progress','completed','archived')),
    region TEXT,
    center_lat FLOAT8,
    center_lng FLOAT8,
    stops_count INT NOT NULL DEFAULT 0,
    entries_count INT NOT NULL DEFAULT 0,
    photos_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);
CREATE INDEX ix_trips_owner ON trips_trips(owner_id, status, updated_at DESC);

CREATE TABLE trips_stops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips_trips(id) ON DELETE CASCADE,
    place_id UUID,
    title TEXT NOT NULL,
    latitude FLOAT8 NOT NULL,
    longitude FLOAT8 NOT NULL,
    planned_at DATE,
    visited_at TIMESTAMPTZ,
    notes TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_stops_trip ON trips_stops(trip_id, sort_order);

CREATE TABLE trips_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips_trips(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    day INT,
    color TEXT NOT NULL DEFAULT '#0E7C7B',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE trips_route_stops (
    route_id UUID NOT NULL REFERENCES trips_routes(id) ON DELETE CASCADE,
    stop_id UUID NOT NULL REFERENCES trips_stops(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    PRIMARY KEY (route_id, stop_id)
);

CREATE TABLE trips_outbox (...);  -- стандартная
```

## 6. Worker: статус-tracker

В `worker/status_tracker.go` cron каждый час:
```sql
UPDATE trips_trips SET status='in_progress'
WHERE status='planned' AND start_date <= CURRENT_DATE AND end_date >= CURRENT_DATE;

UPDATE trips_trips SET status='completed'
WHERE status='in_progress' AND end_date < CURRENT_DATE;
```

И публикуем outbox `trips.trip_started` для notifications.

## 7. События в outbox

- `trips.trip_created`, `trips.trip_started`, `trips.trip_completed` → notifications.
- При создании trip и attached photos обновлять счётчики через journal events.

## 8. Чек-лист

- [ ] CRUD trips, stops, routes
- [ ] Статус-транзишены автоматически по датам
- [ ] Owner-only access enforced на каждом эндпоинте
- [ ] Counters обновляются через internal API из journal/media

## 9. Связанные файлы

- journal → `09_JOURNAL_SERVICE.md`
- places → `07_PLACES_SERVICE.md` (place_id ссылается туда)
- БД → `15_DATABASE_SCHEMAS.md`
