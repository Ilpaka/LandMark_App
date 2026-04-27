# 15 · Полные БД-схемы и миграции

> Все Postgres-схемы LandMark в одном месте, разбитые по сервисам. Каждый сервис применяет только свои миграции через `cmd/migrate/main.go` (goose).

## 1. Общая БД

База данных: **`landmark`**. Хост: `postgres` в Docker-сети.

Создание БД и юзеров — в init-скрипте Postgres (`scripts/init-db.sql`, монтируем в `/docker-entrypoint-initdb.d/`):

```sql
-- scripts/init-db.sql (выполняется один раз при создании контейнера)
CREATE USER auth WITH PASSWORD 'auth';
CREATE USER profile WITH PASSWORD 'profile';
CREATE USER places WITH PASSWORD 'places';
CREATE USER trips WITH PASSWORD 'trips';
CREATE USER journal WITH PASSWORD 'journal';
CREATE USER media WITH PASSWORD 'media';
CREATE USER favorites WITH PASSWORD 'favorites';
CREATE USER notifications WITH PASSWORD 'notifications';
CREATE USER moderation WITH PASSWORD 'moderation';

CREATE SCHEMA auth          AUTHORIZATION auth;
CREATE SCHEMA profile       AUTHORIZATION profile;
CREATE SCHEMA places        AUTHORIZATION places;
CREATE SCHEMA trips         AUTHORIZATION trips;
CREATE SCHEMA journal       AUTHORIZATION journal;
CREATE SCHEMA media         AUTHORIZATION media;
CREATE SCHEMA favorites     AUTHORIZATION favorites;
CREATE SCHEMA notifications AUTHORIZATION notifications;
CREATE SCHEMA moderation    AUTHORIZATION moderation;

-- Pgcrypto доступен всем
CREATE EXTENSION IF NOT EXISTS pgcrypto;

GRANT USAGE ON SCHEMA auth TO auth;
-- ... по аналогии для каждой
```

> В dev (single-tenant) можно использовать одного юзера `landmark` с правами на всё. Production-вариант с разделёнными правами — Phase 3.

## 2. Auth schema (наследуется из Tennis)

> Файлы миграций копируются 1:1 из `Tennis_app/backend/auth-service/db/migrations/`. Только в первой миграции добавляется `CREATE SCHEMA IF NOT EXISTS auth; SET search_path TO auth, public;`.

Таблицы (общая суть):
- `auth_accounts` — пользователи (id, email, status, role, …)
- `auth_credentials` — argon2id-хеш пароля
- `auth_email_verifications` — OTP подтверждения email
- `auth_password_resets` — OTP сброса пароля
- `auth_phone_otps` — phone-flow (есть, но в UI скрыто)
- `auth_sessions` — refresh-токены (с rotation, family, compromised)
- `auth_audit_log` — аудит
- `auth_outbox` — transactional outbox

Полные DDL — в исходниках Tennis. Не копируем сюда чтобы не расходились.

## 3. Profile schema

```sql
-- 20260502120000_init_profile.sql
-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS profile;
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
    user_id UUID,
    expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_reservation_nickname_active
    ON profile_nickname_reservations(nickname_normalized)
    WHERE confirmed_at IS NULL AND expires_at > now();

CREATE TABLE profile_privacy_settings (
    user_id UUID PRIMARY KEY,
    profile_visibility TEXT NOT NULL DEFAULT 'public' CHECK (profile_visibility IN ('public','private')),
    trips_visibility TEXT NOT NULL DEFAULT 'private' CHECK (trips_visibility IN ('public','friends','private')),
    journal_visibility TEXT NOT NULL DEFAULT 'private' CHECK (journal_visibility IN ('friends','private')),
    analytics_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profile_privacy_settings;
DROP TABLE IF EXISTS profile_nickname_reservations;
DROP TABLE IF EXISTS profile_profiles;
-- +goose StatementEnd
```

## 4. Places schema

```sql
-- 20260502130000_init_places.sql
-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS places;
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

CREATE TABLE places_outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);
CREATE INDEX ix_places_outbox_unpublished ON places_outbox(published_at) WHERE published_at IS NULL;
-- +goose StatementEnd

-- +goose Down ...

-- 20260503120000_seed_categories.sql
-- +goose Up
INSERT INTO places.places_categories (slug, title, icon, color, sort_order) VALUES
    ('museum','Музей','museum','#0E7C7B',10),
    ('gastro','Гастро','utensils','#FF6B47',20),
    ('nature','Природа','tree','#7BC9C8',30),
    ('architecture','Архитектура','building','#0E5C5B',40),
    ('viewpoint','Смотровая','eye','#E7A400',50),
    ('park','Парк','leaf','#159594',60);
```

## 5. Trips schema

```sql
-- 20260502140000_init_trips.sql
CREATE SCHEMA IF NOT EXISTS trips;
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

CREATE TABLE trips_outbox (LIKE places.places_outbox INCLUDING ALL);
```

## 6. Journal schema

```sql
-- 20260502150000_init_journal.sql
CREATE SCHEMA IF NOT EXISTS journal;
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

CREATE TABLE journal_outbox (LIKE places.places_outbox INCLUDING ALL);
```

## 7. Media schema

```sql
-- 20260502160000_init_media.sql
CREATE SCHEMA IF NOT EXISTS media;
SET search_path TO media, public;

CREATE TABLE media_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('photo','avatar')),
    mime TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','finalized','expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_media_uploads_owner ON media_uploads(owner_id) WHERE status='pending';

CREATE TABLE media_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    kind TEXT NOT NULL,
    mime TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    object_key TEXT NOT NULL UNIQUE,
    sha256 TEXT,
    status TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('ready','deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_media_owner ON media_media(owner_id, created_at DESC) WHERE status='ready';
```

## 8. Favorites schema

```sql
-- 20260502170000_init_favorites.sql
CREATE SCHEMA IF NOT EXISTS favorites;
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

## 9. Notifications schema

```sql
-- 20260502180000_init_notifications.sql
CREATE SCHEMA IF NOT EXISTS notifications;
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

## 10. Moderation schema

```sql
-- 20260502190000_init_moderation.sql
CREATE SCHEMA IF NOT EXISTS moderation;
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

## 11. Конвенции миграций

- Имя файла: `<timestamp>_<snake_case_name>.sql` (timestamp в формате `YYYYMMDDHHMMSS`).
- В `goose` каждый файл — двусторонний (`+goose Up` / `+goose Down`).
- В каждой миграции в начале `SET search_path TO <schema>, public;`.
- Никогда не модифицируем существующие миграции — только новые.

## 12. Подключение в коде

`DATABASE_URL` для каждого сервиса:
```
postgres://<schema_owner>:<pwd>@postgres:5432/landmark?sslmode=disable&search_path=<schema>,public
```

В Go-коде дополнительно установить `search_path` через `SET search_path` на каждом подключении (pgx pool acquire hook), чтобы исключить кросс-схемные ошибки.

## 13. Связанные файлы

- Каждый сервис → `05..13`
- Docker-compose с томом для БД → `16_DOCKER_INFRA.md`
