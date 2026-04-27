# 03 · Системная архитектура

> Этот файл задаёт **верхнеуровневую архитектуру**: сервисы, потоки запросов, асинхронность, deployment. Детали отдельных сервисов — в `05..14`.

## 1. C4-Level 1 (Context)

```
            ┌──────────────────┐         ┌─────────────────────┐
            │  Mobile App      │  HTTPS  │  Edge / TLS / Nginx │
            │  (Flutter)       ├────────▶│  (опционально)      │
            └──────────────────┘         └──────────┬──────────┘
                                                    │
                                          ┌─────────▼─────────┐
                                          │   API Gateway     │  (Go/Gin)
                                          │   (auth/jwt + lb) │
                                          └─────────┬─────────┘
                                ┌─────────┬─────────┼─────────┬───────────┐
                                ▼         ▼         ▼         ▼           ▼
                              auth     profile   places    trips        ...
                              ▼         ▼         ▼         ▼
                            Postgres + Redis + S3/MinIO + FCM/APNs/SMTP/SMS
```

## 2. C4-Level 2 (Containers) — все сервисы

| Сервис | Назначение | Stateful? | Файл ТЗ |
|---|---|---|---|
| **api-gateway** | Маршрутизация `/v1/*` к сервисам, JWT-валидация, CORS, rate-limit | нет | 14 |
| **auth-service** | Регистрация, OTP-подтверждение, вход, refresh, сессии, аудит | Postgres + Redis | 05 |
| **profile-service** | Никнеймы (резервирование), аватары, био, приватность | Postgres | 06 |
| **places-service** | POI, категории, гео-поиск, статусы (draft/pending/published/rejected) | Postgres + Redis (cache) | 07 |
| **trips-service** | Поездки, точки маршрута, маршруты | Postgres | 08 |
| **journal-service** | Записи журнала, привязка к поездкам и точкам | Postgres | 09 |
| **media-service** | Presigned upload/download, превью | S3/MinIO + Postgres (метаданные) | 10 |
| **favorites-service** | Избранное (POI/Trip/Entry), коллекции | Postgres | 11 |
| **notifications-service** | Push (FCM/APNs), email/SMS, in-app feed, preferences | Postgres + Redis (queue) | 12 |
| **moderation-service** | Очередь модерации публичных POI, действия админа | Postgres | 13 |

Внешние зависимости (managed/контейнерные):
- **Postgres 16** — отдельная БД на каждый сервис, либо отдельные схемы внутри одной БД (см. ниже §6).
- **Redis 7** — общий, namespace-разделение по префиксам ключей.
- **MinIO** — локальная замена S3 в dev/CI.
- **Mailhog** — локальная замена SMTP в dev/CI.
- **FCM/APNs** — продакшен; в dev отключены.
- **OTel collector** + **Tempo** + **Prometheus** + **Grafana** — observability stack (опционально, через профиль `--profile observability`).

## 3. Принципы изоляции и коммуникации

### 3.1 Database-per-service

- Каждый сервис **владеет своими таблицами**. Никаких cross-service JOIN на уровне SQL.
- Чтение данных из чужого сервиса — только через его HTTP API.
- Реализуем как **отдельные схемы Postgres** (`auth.*`, `profile.*`, `places.*`, …) внутри одной БД для упрощения dev-окружения. Каждый сервис подключается с `search_path=<его_схема>` и не имеет grants на другие схемы.

### 3.2 Синхронные вызовы (HTTP)

Используются только когда вызывающий **не может продолжить без данных**. Примеры:
- `auth-service` → `profile-service` при регистрации: резервирование никнейма (см. `ProfileClient` в Tennis-исходнике).
- `api-gateway` → `auth-service` при первом запросе на закрытом эндпоинте: проверка JTI в blacklist (через Redis напрямую быстрее — оставляем Redis).

### 3.3 Асинхронные события (transactional outbox)

Tennis уже содержит паттерн **outbox + Outbox poller**. Используем его:

- `auth.user_registered` → `profile-service` создаёт пустой профиль; `notifications-service` отправляет welcome push.
- `places.place_submitted` → `moderation-service` добавляет в очередь.
- `places.place_published` → `notifications-service` уведомляет автора + кэш-инвалидация.
- `trips.trip_started` (по дате начала, по cron) → `notifications-service` пушит «Поездка началась».

В MVP реализуем самый простой транспорт: **общая таблица outbox в каждом сервисе + Redis Pub/Sub-канал**, на который подписываются интересующиеся сервисы. Без RabbitMQ/Kafka — переусложнение для MVP. Структура канала: `events.<aggregate>.<event_type>`. См. файл 12 для деталей подписки.

### 3.4 Аутентификация запросов

- Клиент → Gateway: `Authorization: Bearer <access_token>` (RS256, выдан auth-service).
- Gateway → внутренний сервис: добавляет HTTP-заголовки:
  - `X-User-Id: <uuid>` — sub из JWT
  - `X-Session-Id: <uuid>`
  - `X-Role: user|admin|guest`
  - `X-Request-Id: <uuid>` — для tracing (если не пришёл от клиента)
- Внутренние сервисы доверяют этим заголовкам **только если** запрос пришёл из приватной сети (см. `14_API_GATEWAY.md` про network isolation в Docker).

## 4. Поток типичного запроса (последовательность)

Пример: «Создание записи журнала с фото»

```
1. Flutter: пользователь нажимает «Сохранить» в JR-03
   → POST /v1/media/uploads (request presigned URL)
   → PUT <S3 presigned> (загружает файл напрямую в MinIO)
   → POST /v1/media/uploads/{id}/finalize (подтверждает загрузку)
   → POST /v1/journal/entries (с media_ids)

2. Gateway:
   - Валидирует JWT (RS256, JWKS из auth-service /jwks)
   - Проверяет, что jti не в blacklist (Redis)
   - Проксирует с заголовками X-User-Id, X-Role

3. journal-service:
   - Валидирует X-User-Id
   - Проверяет существование trip_id (вызов trips-service ИЛИ из локальной denormalized таблицы trip_membership)
   - Создаёт entry, привязывает media_ids
   - Записывает в outbox: journal.entry_created
   - Отвечает 201 с entry_id

4. notifications-service (через Redis Pub/Sub):
   - Получает событие journal.entry_created
   - Решает, нужен ли пуш (нет, для своих записей)
   - Сохраняет in-app уведомление "Запись добавлена"
```

## 5. Технологический стек

### 5.1 Backend

| Компонент | Версия | Зачем |
|---|---|---|
| Go | 1.25 | Язык |
| Gin | v1.12 | HTTP-фреймворк |
| pgx/v5 | v5.7+ | Postgres драйвер |
| go-redis/v9 | v9.7+ | Redis клиент |
| pressly/goose | v3.22+ | Миграции |
| go-jose/v4 | v4.1+ | JWT RS256 |
| google/uuid | v1.6+ | UUID |
| OpenTelemetry SDK | v1.43+ | Трейсинг |
| Prometheus client | v1.20+ | Метрики |
| stretchr/testify | v1.11+ | Юнит-тесты |
| ory/dockertest | v3.11+ | Интеграционные тесты с реальной БД |
| go.uber.org/mock | v0.6+ | Кодогенерация моков |

> Эти версии совпадают с Tennis go.mod — используем как baseline.

### 5.2 Flutter

| Компонент | Версия | Зачем |
|---|---|---|
| Flutter SDK | 3.16+ | UI-фреймворк |
| Dart | 3.4+ | Язык |
| flutter_riverpod | ^2.6 | State management + DI |
| go_router | ^14.0 | Декларативный роутинг |
| dio | ^5.7 | HTTP-клиент с interceptors |
| freezed | ^2.5 | Immutable models, sum-types для UI state |
| json_serializable | ^6.8 | JSON ↔ Dart |
| flutter_secure_storage | ^9.2 | Хранение access/refresh-токенов |
| drift | ^2.20 | Локальная SQLite БД (offline) |
| cached_network_image | ^3.4 | Кэш изображений |
| flutter_map | ^7.0 | OpenStreetMap-движок (без Google billing) |
| latlong2 | ^0.9 | Гео-типы |
| geolocator | ^12.0 | Геопозиция устройства |
| permission_handler | ^11.0 | PR-01..03 |
| firebase_messaging | ^15.0 | FCM пушей |
| image_picker | ^1.1 | JR-06, MP-05 |
| photo_view | ^0.15 | JR-05 |

> Полный pubspec — в `19_FLUTTER_DESIGN_SYSTEM.md` секции «Зависимости».

## 6. Стратегия БД

### 6.1 Один Postgres-кластер, схемы по сервисам

```
landmark_db
├── auth.*       — auth-service (наследует Tennis)
├── profile.*    — profile-service
├── places.*     — places-service
├── trips.*      — trips-service
├── journal.*    — journal-service
├── media.*      — media-service
├── favorites.*  — favorites-service
├── notifications.* — notifications-service
└── moderation.* — moderation-service
```

Каждый сервис подключается под своим Postgres-юзером с `GRANT USAGE ON SCHEMA <его_схема>` и без прав на чужие. Это предотвращает случайные cross-service JOIN'ы и упрощает потом разбить БД физически.

### 6.2 Connection strings

Сервис получает `DATABASE_URL` через env. Паттерн в docker-compose:
```
DATABASE_URL=postgres://auth:auth@postgres:5432/landmark?search_path=auth
```

### 6.3 Миграции

- Каждый сервис содержит `db/migrations/<timestamp>_<name>.sql` (формат goose).
- Каждый сервис содержит `cmd/migrate/main.go`, копирует структуру из Tennis.
- Миграции применяются **до запуска сервера** через `docker-entrypoint.sh` (как в Tennis).

## 7. Безопасность на архитектурном уровне

- **Внешний периметр**: только api-gateway открыт наружу (`8080`); все остальные сервисы слушают `127.0.0.1`/в Docker-сети `landmark-internal`.
- **TLS**: вне Docker — terminate на прокси/edge (Nginx/ALB). Для dev — без TLS.
- **JWT**: RS256, key rotation поддерживается через JWKS endpoint в auth-service (унаследовано).
- **Refresh-токены**: только в `auth.auth_sessions`, хешируются (унаследовано).
- **Pepper**: `AUTH_PEPPER` — секрет для argon2id; передаётся через env, не коммитится.
- **Все чувствительные данные** маскируются в логах (см. `audit/mask.go` Tennis — копируется как есть).

## 8. Развёртывание (deployment)

### 8.1 Local dev

```
docker compose up -d        # все сервисы + postgres + redis + minio + mailhog
docker compose logs -f auth-service
```

Должно работать с zero-config: первый старт делает `openssl genrsa` для JWT, миграции и стартует сервер.

### 8.2 CI

- На каждый push в `create_app`: `make test`, `make test-integration` (через dockertest).
- На каждый push в `main`: build всех образов и публикация в локальный registry (опционально).

### 8.3 Prod (вне scope MVP, но проектируем под)

- Один Docker host или k8s. Каждый сервис — один deployment.
- Postgres — managed (e.g., AWS RDS).
- Redis — managed (e.g., ElastiCache).
- S3 — managed.
- Секреты — через `secrets manager` или k8s Secret.

## 9. Связанные файлы

- Структура каталогов и общие пакеты Go → `04_BACKEND_LAYOUT.md`
- API-Gateway → `14_API_GATEWAY.md`
- БД-схемы → `15_DATABASE_SCHEMAS.md`
- Docker → `16_DOCKER_INFRA.md`
- Observability и security → `22_OBSERVABILITY_SECURITY.md`
