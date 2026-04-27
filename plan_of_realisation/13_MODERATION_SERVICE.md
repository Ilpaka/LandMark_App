# 13 · moderation-service

> Очередь модерации публичных POI и действий администратора (одобрить/отклонить).

## 1. Bounded Context

moderation-service владеет: queue items, decisions, moderator stats. Ссылается на `place_id` (внешний агрегат places).

## 2. Поток

```
[user] POST /v1/places/{id}/submit
   places-service: status → pending_moderation, outbox places.place_submitted

[moderation-service worker] subscribes events.places_place.place_submitted:
   создаёт moderation_items {place_id, status=open, submitted_at, author_id}

[admin] открывает AD-02 (очередь)
   GET /v1/moderation/queue?cursor=&limit=

[admin] открывает AD-03 (карточка) → решение
   POST /v1/moderation/items/{id}/decision { action: approve | reject, reason? }
   moderation-service:
     1. Меняет item.status (approved | rejected) и записывает decision
     2. Вызывает places-service internal:
        - approve → places-service публикует place (status=published)
        - reject → places-service выставляет status=rejected + reason
     3. places-service публикует outbox place_published / place_rejected
     4. notifications-service узнаёт об этом и уведомляет автора
```

## 3. Сущности

```go
type Item struct {
    ID          uuid.UUID
    PlaceID     uuid.UUID
    AuthorID    uuid.UUID
    Status      string  // open | approved | rejected
    Priority    int     // 0..2 (приоритет очереди, 0=normal)
    SubmittedAt time.Time
    AssignedTo  *uuid.UUID
    DecidedAt   *time.Time
    DecidedBy   *uuid.UUID
    Decision    *string  // approve | reject
    Reason      *string
}

type Stat struct {
    AdminID   uuid.UUID
    Date      time.Time  // truncated to day
    Approved  int
    Rejected  int
}
```

## 4. Use cases

- `ListQueue(ctx, status, cursor, limit)` — AD-02. По умолчанию `status=open`, sort `submitted_at ASC` (FIFO).
- `Get(ctx, itemID)` — AD-03. Включает поля POI через places-service GET.
- `Decide(ctx, adminID, itemID, action, reason)` — AD-03 «Да/Нет». Транзакционно: пишет decision, вызывает places-service internal, обновляет stats. На failure — откат.
- `Stats(ctx, adminID, period)` — AD-01.
- `ReassignAll(ctx, fromAdminID, toAdminID)` — для будущего.

## 5. HTTP API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| GET | `/v1/moderation/queue` | admin | `?status=&cursor=&limit=` (AD-02) |
| GET | `/v1/moderation/items/{id}` | admin | (AD-03) |
| POST | `/v1/moderation/items/{id}/assign` | admin | self-assign |
| POST | `/v1/moderation/items/{id}/decision` | admin | `{action: approve|reject, reason?}` |
| GET | `/v1/moderation/stats` | admin | `?period=day|week|month` (AD-01) |

## 6. БД-схема

```sql
SET search_path TO moderation, public;

CREATE TABLE moderation_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    place_id UUID NOT NULL UNIQUE,
    author_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','approved','rejected')),
    priority SMALLINT NOT NULL DEFAULT 0,
    assigned_to UUID,
    submitted_at TIMESTAMPTZ NOT NULL,
    decided_at TIMESTAMPTZ,
    decided_by UUID,
    decision TEXT CHECK (decision IS NULL OR decision IN ('approve','reject')),
    reason TEXT
);
CREATE INDEX ix_mod_open ON moderation_items(status, priority DESC, submitted_at ASC) WHERE status='open';

CREATE TABLE moderation_stats_daily (
    admin_id UUID NOT NULL,
    day DATE NOT NULL,
    approved INT NOT NULL DEFAULT 0,
    rejected INT NOT NULL DEFAULT 0,
    PRIMARY KEY (admin_id, day)
);
```

## 7. Чек-лист

- [ ] Subscriber на `places.place_submitted` создаёт row с `status=open`
- [ ] Decision требует `role=admin` (через JWT)
- [ ] При decision=reject `reason` обязателен
- [ ] places-service internal-вызов идемпотентен (повторное approve того же — 200)
- [ ] Stats обновляются атомарно в той же транзакции

## 8. Связанные файлы

- places → `07_PLACES_SERVICE.md` (internal endpoints approve/reject)
- notifications → `12_NOTIFICATIONS_SERVICE.md`
- БД → `15`
