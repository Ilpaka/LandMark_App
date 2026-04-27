# LandMark (Wanderlog) — Технический план реализации

> Полный набор технических документов для реализации мобильного приложения «Журнал путешественника с картой достопримечательностей». ИИ-исполнитель должен последовательно проработать каждый файл, перекрёстно сверяясь с другими.

## 0. Контекст

| Параметр | Значение |
|---|---|
| Кодовое имя | **Wanderlog** (внутренний продукт LandMark) |
| Платформы клиента | iOS + Android (одна Flutter-кодовая база) |
| Стек бэкенда | Go 1.25 + Gin (HTTP), PostgreSQL 16, Redis 7, S3/MinIO, OpenTelemetry |
| Стек клиента | Flutter 3.x (Dart 3.x), Material 3, Riverpod, GoRouter, Dio |
| Контейнеризация | Docker + docker-compose (single root compose orchestrates all services) |
| Архитектура | Clean Architecture (Hex/Onion), per-service Bounded Context |
| Базовый шаблон | Auth-сервис скопирован из `/Users/ilpaka/Development/Tennis_app/backend/auth-service` и адаптирован |

## 1. Структура каталогов проекта

```
LandMark_App/
├── backend/                 # Все Go-сервисы (см. файлы 04..14)
│   ├── api-gateway/         # File 14
│   ├── auth-service/        # File 05 (адаптация Tennis)
│   ├── profile-service/     # File 06
│   ├── places-service/      # File 07
│   ├── trips-service/       # File 08
│   ├── journal-service/     # File 09
│   ├── media-service/       # File 10
│   ├── favorites-service/   # File 11
│   ├── notifications-service/ # File 12
│   ├── moderation-service/  # File 13
│   ├── pkg/                 # Shared packages (observability, httpx, ...)
│   └── docker-compose.yml   # File 16 (root compose)
├── landmark_app/            # Flutter (см. файлы 18..20)
│   └── lib/
│       ├── core/
│       └── features/
└── plan_of_realisation/     # ← вы здесь
```

## 2. Дорожная карта файлов

| # | Файл | Назначение | Кому |
|---|---|---|---|
| 00 | [00_INDEX.md](00_INDEX.md) | Этот файл | Все |
| 01 | [01_PRODUCT.md](01_PRODUCT.md) | Роли, ключевые сценарии, MVP-границы, нефункциональные требования | PM/архитектор |
| 02 | [02_SCREENS.md](02_SCREENS.md) | Каталог 68 экранов с привязкой к API/состояниям | Flutter + Backend |
| 03 | [03_ARCHITECTURE.md](03_ARCHITECTURE.md) | Системная архитектура: сервисы, потоки, асинхронность, deployment | Архитектор |
| 04 | [04_BACKEND_LAYOUT.md](04_BACKEND_LAYOUT.md) | Принципы Clean Architecture в Go, layout сервиса, общие пакеты | Backend |
| 05 | [05_AUTH_SERVICE.md](05_AUTH_SERVICE.md) | Авторизация: что копируем 1:1 из Tennis, что меняем под Wanderlog | Backend |
| 06 | [06_PROFILE_SERVICE.md](06_PROFILE_SERVICE.md) | Профиль пользователя, никнейм, аватар, приватность | Backend |
| 07 | [07_PLACES_SERVICE.md](07_PLACES_SERVICE.md) | Достопримечательности (POI), категории, гео-поиск, статусы модерации | Backend |
| 08 | [08_TRIPS_SERVICE.md](08_TRIPS_SERVICE.md) | Поездки: создание, точки, привязка к POI, обложки | Backend |
| 09 | [09_JOURNAL_SERVICE.md](09_JOURNAL_SERVICE.md) | Записи журнала, реакции, привязка к точкам поездки | Backend |
| 10 | [10_MEDIA_SERVICE.md](10_MEDIA_SERVICE.md) | Загрузка/выдача медиа, presigned URL, S3/MinIO | Backend |
| 11 | [11_FAVORITES_SERVICE.md](11_FAVORITES_SERVICE.md) | Избранное (POI/Trips/Journal), коллекции | Backend |
| 12 | [12_NOTIFICATIONS_SERVICE.md](12_NOTIFICATIONS_SERVICE.md) | Push (FCM/APNs), email, SMS, in-app уведомления | Backend |
| 13 | [13_MODERATION_SERVICE.md](13_MODERATION_SERVICE.md) | Очередь модерации публичных POI, действия админа | Backend |
| 14 | [14_API_GATEWAY.md](14_API_GATEWAY.md) | API Gateway: маршрутизация, JWT-проверка, CORS, rate-limit | Backend |
| 15 | [15_DATABASE_SCHEMAS.md](15_DATABASE_SCHEMAS.md) | Все Postgres-схемы и миграции goose, разделённые по сервисам | Backend |
| 16 | [16_DOCKER_INFRA.md](16_DOCKER_INFRA.md) | Корневой docker-compose, Dockerfile-шаблон, env-файлы, CI | DevOps |
| 17 | [17_API_CONTRACTS.md](17_API_CONTRACTS.md) | Каркас OpenAPI 3.0 на каждый сервис + единые error/pagination модели | Backend |
| 18 | [18_FLUTTER_ARCHITECTURE.md](18_FLUTTER_ARCHITECTURE.md) | Clean Architecture в Flutter: layers, DI, networking, state | Flutter |
| 19 | [19_FLUTTER_DESIGN_SYSTEM.md](19_FLUTTER_DESIGN_SYSTEM.md) | Дизайн-токены, темы, компоненты по wanderlog_screens.html | Flutter |
| 20 | [20_FLUTTER_FEATURES.md](20_FLUTTER_FEATURES.md) | Feature-модули, маппинг экранов на BLoC/use-cases | Flutter |
| 21 | [21_TESTING.md](21_TESTING.md) | Пирамида тестов: unit, integration, e2e, contract | QA |
| 22 | [22_OBSERVABILITY_SECURITY.md](22_OBSERVABILITY_SECURITY.md) | Логи, метрики, трейсинг, security baseline | DevOps |
| 23 | [23_ROADMAP.md](23_ROADMAP.md) | Фазы реализации с приёмочными критериями (DoD) | Все |

## 3. Принцип чтения для ИИ-исполнителя

1. **Перед стартом любой задачи** прочитай `01_PRODUCT.md` и `03_ARCHITECTURE.md` целиком.
2. **Перед тем как писать сервис X** — прочитай файл `05..14` для X, потом `15` (его таблицы), `17` (его API), затем `04` (общие принципы).
3. **Перед тем как писать клиентский экран** — прочитай его строку в `02_SCREENS.md`, затем `19` (компоненты), `18` (архитектура), `20` (фичи).
4. **Никогда не отступай от структуры папок и неймингов из этого ТЗ** — вся coheren­тность проекта на ней держится.
5. **Если что-то не описано** — выбери самый простой вариант, не нарушающий принципы Clean Architecture, и зафиксируй его в коде однострочным комментарием формата `// DECISION: <что и почему>`.

## 4. Базовые источники истины

- Дизайн всех экранов: `WIKI/design/wanderlog_screens.html` — единый источник по UI/UX, цвет, типографика, состояния, иконография.
- Спецификация: `WIKI/design/wanderlog_specification.docx` — детали взаимодействия и пользовательских историй.
- Этапы учебной практики (`WIKI/Stage_1..7`) — обоснование архитектурных решений, диаграммы (as-is/to-be, классы, ассоциации, декомпозиция).
- Базовый сервис авторизации: `Tennis_app/backend/auth-service` — копируется и адаптируется (см. `05_AUTH_SERVICE.md`).

## 5. Соглашения по версионированию и API

- HTTP-API сервисов: префикс `/v1/...`. Версионирование вперёд только через `/v2/...`, без break внутри `/v1/`.
- Имена ресурсов: множественное, kebab-case в URL (`/v1/places`, `/v1/trips/{id}/journal-entries`).
- Идентификаторы: всегда `UUID v4`, в БД — тип `UUID`, в JSON — строка.
- Время: `TIMESTAMPTZ` в БД, `RFC3339Nano` в JSON.
- Координаты: `latitude FLOAT8 NOT NULL CHECK (latitude BETWEEN -90 AND 90)`, `longitude FLOAT8 NOT NULL CHECK (longitude BETWEEN -180 AND 180)`. В Postgres гео-индексы: `PostGIS` отключаем — используем `earthdistance` + GiST для ускорения; если в будущем нужен сложный гео-поиск, мигрируем на PostGIS отдельной миграцией.
- Пагинация: только cursor-based (`?cursor=<base64>&limit=<int>`), никаких offset.
- Ошибки: единый формат `ErrorBody { error: string, code?: string, details?: object }` (см. `17_API_CONTRACTS.md`).

## 6. Чек-лист готовности к Phase 1 (см. `23_ROADMAP.md`)

- [ ] Создан `backend/` каталог с подкаталогом каждого сервиса
- [ ] Создан `backend/pkg/observability` со всеми отонесёнными зависимостями (см. `04`)
- [ ] auth-service запускается через `docker compose up auth-service` (см. `05` + `16`)
- [ ] api-gateway маршрутизирует `/v1/auth/*` в auth-service (см. `14`)
- [ ] Flutter проект собирается: `flutter run` показывает экран Splash → Welcome (см. `20`)
- [ ] Полная цепочка регистрация → подтверждение OTP → вход → `/v1/auth/me` работает end-to-end
