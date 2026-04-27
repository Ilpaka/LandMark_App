# 12 · notifications-service

> In-app feed уведомлений + отправка через каналы (push FCM/APNs, email, SMS).

## 1. Bounded Context

notifications-service владеет: notifications (in-app), devices (push tokens), preferences (что доставлять).

Получает события от других сервисов через Redis Pub/Sub (см. `pkg/eventbus`).

## 2. Сущности

```go
type Notification struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Type      NotificationType  // см. ниже
    Title     string
    Body      string
    DeepLink  *string           // например "/main/places/{id}"
    Meta      map[string]any    // structured data
    ReadAt    *time.Time
    CreatedAt time.Time
}

type NotificationType string  // place_approved | place_rejected | trip_started | sync_done | welcome | password_changed

type Device struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    Platform   string   // ios | android
    PushToken  string   // FCM / APNs token
    LastSeenAt time.Time
    CreatedAt  time.Time
}

type Preference struct {
    UserID                uuid.UUID
    PushTripReminders     bool
    PushModerationResult  bool
    PushSyncStatus        bool
    PushMarketing         bool
    EmailWelcome          bool
    EmailModerationResult bool
    EmailMarketing        bool
    UpdatedAt             time.Time
}
```

## 3. Use cases

- `RegisterDevice(ctx, userID, platform, token)` — после согласия на пуши (PR-02).
- `UnregisterDevice(ctx, userID, deviceID)` — на logout.
- `GetPreferences(ctx, userID)` — PF-04.
- `UpdatePreferences(ctx, userID, p)`.
- `ListInApp(ctx, userID, cursor, limit)` — NT-01.
- `MarkRead(ctx, userID, notifIDs)`.
- Internal: `Enqueue(ctx, userID, type, title, body, deepLink, meta)` — вызывается обработчиком event subscriber'a.

## 4. HTTP API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| GET | `/v1/notifications` | user | `?cursor=&limit=` (NT-01) |
| POST | `/v1/notifications/read` | user | `{ids: []}` |
| POST | `/v1/notifications/devices` | user | `{platform, token}` |
| DELETE | `/v1/notifications/devices/{id}` | user | удалить |
| GET | `/v1/notifications/preferences` | user | PF-04 |
| PATCH | `/v1/notifications/preferences` | user | |
| POST | `/v1/notifications/internal/enqueue` | internal | для других сервисов (если событие нужно срочно, без eventbus) |

## 5. БД-схема

```sql
SET search_path TO notifications, public;

CREATE TABLE notifications_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    deep_link TEXT,
    meta JSONB NOT NULL DEFAULT '{}',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_notif_user_unread ON notifications_notifications(user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX ix_notif_user_all ON notifications_notifications(user_id, created_at DESC);

CREATE TABLE notifications_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('ios','android')),
    push_token TEXT NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_device_token ON notifications_devices(push_token);

CREATE TABLE notifications_preferences (
    user_id UUID PRIMARY KEY,
    push_trip_reminders BOOLEAN NOT NULL DEFAULT TRUE,
    push_moderation_result BOOLEAN NOT NULL DEFAULT TRUE,
    push_sync_status BOOLEAN NOT NULL DEFAULT FALSE,
    push_marketing BOOLEAN NOT NULL DEFAULT FALSE,
    email_welcome BOOLEAN NOT NULL DEFAULT TRUE,
    email_moderation_result BOOLEAN NOT NULL DEFAULT TRUE,
    email_marketing BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 6. Подписки на события

```go
// worker/subscriber.go
sub.Subscribe(ctx, "events.auth_account.user_registered", handleUserRegistered)
sub.Subscribe(ctx, "events.places_place.place_published", handlePlacePublished)
sub.Subscribe(ctx, "events.places_place.place_rejected", handlePlaceRejected)
sub.Subscribe(ctx, "events.trips_trip.trip_started", handleTripStarted)
sub.Subscribe(ctx, "events.trips_trip.trip_completed", handleTripCompleted)
```

Каждый handler:
1. Читает preferences пользователя
2. Создаёт in-app notification
3. Если push разрешён — отправляет через FCM/APNs
4. Если email разрешён — отправляет через SMTP

## 7. Адаптеры внешних каналов

- `EmailSender` — SMTP (с Mailhog в dev, реальный SMTP в prod). Унаследован из Tennis (`adapters/email/sender.go`).
- `PushSender` — единый интерфейс, имплементации: FCM (HTTP v1 API) для Android, APNs (token-based) для iOS. Mock в dev-режиме.
- `SMSSender` — унаследован из Tennis (`adapters/sms/{log,smsru}.go`).

## 8. Шаблоны уведомлений

В коде (`internal/modules/notification/templates/`):
- `welcome.go` — после регистрации
- `place_approved.go` — «Ваше место «X» одобрено!»
- `place_rejected.go` — с reason
- `trip_started.go` — «Поездка «X» началась»
- ...

Тексты — в виде Go-шаблонов с подстановками. Локализация — только `ru` в MVP.

## 9. Чек-лист

- [ ] Регистрация устройства идемпотентна по token
- [ ] Подписка на eventbus поднимается в worker'е
- [ ] In-app feed имеет cursor-based pagination
- [ ] Mark-read меняет only own notifications
- [ ] При выключенных preferences пуш не уходит (но in-app всё равно создаётся)

## 10. Связанные файлы

- eventbus → `04_BACKEND_LAYOUT.md` §4.4
- email/sms → шаблоны Tennis
- БД → `15`
