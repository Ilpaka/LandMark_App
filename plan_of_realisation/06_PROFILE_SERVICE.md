# 06 · profile-service

> Хранит данные профиля пользователя, никнеймы (с unique-резервированием), аватары, биографию, настройки приватности.

## 1. Bounded Context

profile-service владеет: profile, nickname-reservation, privacy-settings, blocked-users (для будущего social).

profile-service **не** владеет: email/password (это auth), аватары как файлы (это media-service — здесь храним только `avatar_media_id`).

## 2. Структура каталогов

См. шаблон в `04_BACKEND_LAYOUT.md`. Модуль — `internal/modules/profile/`.

## 3. Доменные сущности

```go
// domain/profile.go
type Profile struct {
    UserID       uuid.UUID
    Nickname     string         // уникальный, lowercase, [a-z0-9_], 3..32
    DisplayName  string         // human-readable, 1..64
    Bio          *string        // до 500 chars
    AvatarMediaID *uuid.UUID    // ссылка на media-service
    City         *string
    Country      *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// domain/privacy.go
type PrivacySettings struct {
    UserID            uuid.UUID
    ProfileVisibility Visibility  // public | private
    TripsVisibility   Visibility  // public | friends | private
    JournalVisibility Visibility  // friends | private (нет public в MVP)
    AnalyticsEnabled  bool
    UpdatedAt         time.Time
}
type Visibility string  // const: public, friends, private

// domain/nickname_reservation.go
type NicknameReservation struct {
    ID         uuid.UUID
    Nickname   string
    UserID     *uuid.UUID  // null до confirm
    ExpiresAt  time.Time   // TTL 5 минут
    ConfirmedAt *time.Time
    CreatedAt  time.Time
}
```

## 4. Ports (interfaces)

```go
// ports/ports.go
type Store interface {
    WithTx(ctx, fn) error

    GetProfileByUserID(ctx, tx, uid uuid.UUID) (*Profile, error)
    GetProfileByNickname(ctx, tx, nickname string) (*Profile, error)
    InsertProfile(ctx, tx, p Profile) error
    UpdateProfile(ctx, tx, p Profile) error

    GetPrivacy(ctx, tx, uid uuid.UUID) (*PrivacySettings, error)
    UpsertPrivacy(ctx, tx, ps PrivacySettings) error

    InsertReservation(ctx, tx, nick string) (*NicknameReservation, error)
    GetReservation(ctx, tx, id uuid.UUID) (*NicknameReservation, error)
    ConfirmReservation(ctx, tx, id, userID uuid.UUID) error
    DeleteReservation(ctx, tx, id uuid.UUID) error
    DeleteExpiredReservations(ctx, tx) error
}

type MediaClient interface {  // для проверки существования media и его владельца
    GetMediaOwner(ctx, mediaID uuid.UUID) (uuid.UUID, error)
}
```

## 5. Use cases (Service)

- `ReserveNickname(ctx, nickname) (reservationID, error)` — `INSERT ... ON CONFLICT DO NOTHING`. Возвращает 409 если уже занят.
- `ConfirmReservation(ctx, reservationID, userID)` — после `verify-email` в auth-service.
- `ReleaseReservation(ctx, reservationID)` — если регистрация не дошла до confirm.
- `GetMyProfile(ctx, userID)` — для PF-01.
- `UpdateMyProfile(ctx, userID, displayName, bio, avatarMediaID, city, country)` — PF-02. Если меняется `avatarMediaID`, проверяем владельца через MediaClient.
- `ChangeNickname(ctx, userID, newNickname)` — резервирование + миграция (для будущего; в MVP отключён).
- `GetPrivacy(ctx, userID)` / `UpdatePrivacy(ctx, userID, settings)` — PF-05.

## 6. HTTP API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| GET | `/v1/profile/me` | user | свой профиль (PF-01) |
| PATCH | `/v1/profile/me` | user | обновить (PF-02). Body partial. |
| GET | `/v1/profile/by-nickname/{nick}` | user | публичный просмотр чужого профиля (для будущего) |
| GET | `/v1/profile/me/privacy` | user | настройки приватности |
| PATCH | `/v1/profile/me/privacy` | user | обновить |
| POST | `/v1/profile/internal/nicknames/reserve` | internal | резервация (вызывает auth) |
| POST | `/v1/profile/internal/nicknames/{rid}/confirm` | internal | подтверждение |
| DELETE | `/v1/profile/internal/nicknames/{rid}` | internal | релиз |
| GET | `/v1/profile/internal/{user_id}` | internal | профиль по user_id (используется другими сервисами) |

`/internal/*` защищён middleware, который проверяет header `X-Internal-Key: <INTERNAL_API_KEY>` и принимает только из приватной Docker-сети (по client IP).

## 7. БД-схема (см. `15_DATABASE_SCHEMAS.md` для полных миграций)

```sql
SET search_path TO profile, public;

CREATE TABLE profile_profiles (
    user_id UUID PRIMARY KEY,
    nickname TEXT NOT NULL,
    nickname_normalized TEXT NOT NULL,
    display_name TEXT NOT NULL,
    bio TEXT,
    avatar_media_id UUID,
    city TEXT,
    country TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_profile_nickname_norm ON profile_profiles(nickname_normalized);

CREATE TABLE profile_nickname_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname_normalized TEXT NOT NULL,
    user_id UUID,                       -- null до confirm
    expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_reservation_nickname_active
    ON profile_nickname_reservations(nickname_normalized)
    WHERE confirmed_at IS NULL AND expires_at > now();

CREATE TABLE profile_privacy_settings (
    user_id UUID PRIMARY KEY,
    profile_visibility TEXT NOT NULL DEFAULT 'public',
    trips_visibility TEXT NOT NULL DEFAULT 'private',
    journal_visibility TEXT NOT NULL DEFAULT 'private',
    analytics_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (profile_visibility IN ('public','private')),
    CHECK (trips_visibility IN ('public','friends','private')),
    CHECK (journal_visibility IN ('friends','private'))
);
```

## 8. Outbox

Не нужен в MVP (нет интересных downstream-событий). При появлении ленты/feed добавим `profile.profile_updated`.

## 9. Чек-лист

- [ ] `POST /v1/profile/internal/nicknames/reserve` возвращает 409 на дубликат
- [ ] `auth-service` успешно вызывает reserve→confirm в registration flow
- [ ] `PATCH /v1/profile/me` валидирует длины и возвращает обновлённый профиль
- [ ] Background-cleanup expired reservations (cron 1 раз в минуту через worker)

## 10. Связанные файлы

- Auth-сервис вызывает profile через `ProfileClient` → см. `05_AUTH_SERVICE.md` Шаг 10
- БД-схема → `15_DATABASE_SCHEMAS.md`
- API → `17_API_CONTRACTS.md`
