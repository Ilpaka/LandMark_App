# 23 · Дорожная карта реализации

> Этапы (фазы) с конкретными приёмочными критериями (DoD). Идти строго по порядку — каждый последующий этап опирается на предыдущий.

## Phase 0 — Подготовка (день 0)

**Цель**: создан скелет репозитория, ветка, минимальная инфра.

Действия:
1. Создать ветку `create_app` (✅ уже сделано).
2. Создать каталоги:
   ```
   backend/
   backend/pkg/observability/
   backend/api-gateway/
   backend/auth-service/
   backend/profile-service/
   backend/places-service/
   backend/trips-service/
   backend/journal-service/
   backend/media-service/
   backend/favorites-service/
   backend/notifications-service/
   backend/moderation-service/
   landmark_app/  (уже есть)
   plan_of_realisation/  (✅ есть, текущий ТЗ)
   ```
3. Скопировать `Tennis_app/backend/pkg/observability/` → `LandMark_App/backend/pkg/observability/` и заменить module path.
4. Скопировать `Tennis_app/backend/auth-service/` → `LandMark_App/backend/auth-service/` целиком.
5. Создать корневой `backend/docker-compose.yml` со скелетом + Postgres + Redis + MinIO + Mailhog.
6. Создать `backend/scripts/init-db.sql` (см. `15_DATABASE_SCHEMAS.md` §1).

**DoD Phase 0:**
- [ ] `docker compose -f backend/docker-compose.yml up postgres redis minio mailhog -d` поднимается без ошибок
- [ ] `psql ... -c '\dn'` показывает все 9 схем
- [ ] `mc ls local/` показывает bucket `landmark-media`

## Phase 1 — auth-service end-to-end

**Цель**: пользователь может зарегистрироваться, подтвердить email, войти, запросить `/me`, выйти.

Действия (см. `05_AUTH_SERVICE.md` детально):
1. Адаптация скопированного auth-service (module path, схема `auth`, новый `/v1/auth/register`).
2. Сборка и запуск через `docker compose up auth-service`.
3. Прохождение всех unit + integration тестов Tennis в новом репозитории.
4. Написание нового теста `register_test.go` для email-flow.
5. Подключение к Mailhog для просмотра OTP в dev.

**DoD Phase 1:**
- [ ] `make test` зелёный
- [ ] `make test-integration` зелёный
- [ ] curl-сценарий: register → достать OTP из Mailhog → verify → login → me → refresh → logout, всё успешно
- [ ] `GET /v1/auth/jwks` отдаёт валидный JWKS

## Phase 2 — Skeleton всех остальных сервисов

**Цель**: каждый сервис собирается, миграции применяются, `/healthz` отвечает 200.

Действия:
1. По шаблону из `04_BACKEND_LAYOUT.md` создать `backend/<service>/` с:
   - `cmd/server/main.go`
   - `cmd/migrate/main.go`
   - `internal/config/config.go`
   - `internal/httpserver/server.go` (с минимальным `/healthz`)
   - `Dockerfile`, `docker-entrypoint.sh`, `Makefile`, `go.mod`
   - `db/migrations/<timestamp>_init_<service>.sql` (DDL из `15_DATABASE_SCHEMAS.md`)
2. Подключить к корневому compose.
3. Проверить, что миграции применяются и каждый `/healthz` отдаёт 200.

**DoD Phase 2:**
- [ ] `docker compose up -d --build` поднимает все сервисы
- [ ] `for s in auth profile places trips journal media favorites notifications moderation; do curl http://${s}-service:8080/healthz; done` все 200 (если внутри сети)
- [ ] У каждого сервиса в БД появились его таблицы

## Phase 3 — api-gateway + JWT-flow

**Цель**: единая точка входа, валидация JWT, маршрутизация в auth.

Действия:
1. Создать `backend/api-gateway/` (см. `14_API_GATEWAY.md`).
2. Реализовать JWT-middleware (через JWKS-кэш).
3. Прописать routing-таблицу для auth + healthz.
4. Подключить Redis для rate-limit и blacklist.
5. Скрыть auth-service за gateway (убрать `ports:` у auth, оставить только у gateway).

**DoD Phase 3:**
- [ ] `curl http://localhost:8080/v1/auth/me -H "Authorization: Bearer <token>"` возвращает корректный профиль
- [ ] `curl http://localhost:8080/v1/auth/me` без токена возвращает 401
- [ ] При истёкшем токене gateway возвращает 401, не проксируя
- [ ] Rate-limit срабатывает на 21-м login за минуту

## Phase 4 — Flutter: auth-flow

**Цель**: установленный мобильный клиент проводит пользователя от splash до главной вкладки.

Действия (см. `18..20`):
1. Перестроить `landmark_app/lib/`:
   - `bootstrap/`, `core/design/`, `core/networking/`, `core/routing/`, `core/storage/`
   - `features/splash/`, `features/onboarding/`, `features/auth/`
2. Подключить зависимости из `20_FLUTTER_FEATURES.md` §1.
3. Реализовать SP-01..04, AU-01..06.
4. ApiClient + AuthInterceptor с refresh.
5. Тестовый прогон на эмуляторе iOS/Android против `http://10.0.2.2:8080`.

**DoD Phase 4:**
- [ ] `flutter run` собирается, splash → welcome → register
- [ ] OTP подтверждение работает
- [ ] После логина видна заглушка вкладки «Карта» с приветствием по имени
- [ ] При killed-app сессия восстанавливается через access/refresh
- [ ] Logout → snapшот state очищается

## Phase 5 — places-service + map (MP-01..04, MP-08)

**Цель**: на карте видны публичные POI, можно открыть карточку.

Действия:
1. Backend places: реализовать `GET /v1/places`, `GET /v1/places/{id}`, `GET /v1/places/categories`. Seed-данные.
2. Прописать маршруты в gateway.
3. Flutter `features/map/`: MapPage, FilterSheet, PlaceDetailPage, PinPreviewSheet.
4. Поиск с подсказками — пока ILIKE по title/city.

**DoD Phase 5:**
- [ ] На карте показываются 6+ seed-точек
- [ ] Tap по пину открывает превью
- [ ] Открывается полная карточка POI с фото (если есть)
- [ ] Фильтр по категории работает

## Phase 6 — trips + journal (TR-*, JR-*) + media (uploads)

**Цель**: пользователь может создать поездку, добавить запись с фото.

Действия:
1. Backend trips: CRUD, stops, statuses (без routes).
2. Backend journal: CRUD entries, attach media.
3. Backend media: presigned upload + finalize.
4. Flutter trips/journal/media фичи.

**DoD Phase 6:**
- [ ] Создание поездки → отображается в TR-01
- [ ] Создание записи с 3 фото → отображается в JR-01
- [ ] Открытие записи → фото подгружаются (через presigned GET)

## Phase 7 — favorites + предложение POI + moderation

**Цель**: полный цикл community contribution.

Действия:
1. Backend favorites: CRUD.
2. Backend places: `POST /v1/places`, `POST /v1/places/{id}/submit`, internal approve/reject.
3. Backend moderation: queue, decision, integration с places.
4. Flutter MP-05..07 + AD-01..04 (admin tab появляется при role=admin).

**DoD Phase 7:**
- [ ] Пользователь может предложить POI
- [ ] В админке (на тестовом аккаунте с `role=admin`) видна очередь
- [ ] Approve/Reject меняет статус и шлёт push автору

## Phase 8 — notifications, settings, profile

**Цель**: все экраны NT-*, PF-*.

Действия:
1. Backend profile: CRUD profile + privacy.
2. Backend notifications: in-app feed, devices, preferences, FCM-интеграция.
3. Flutter PF-01..08 + NT-01..03 + PR-01..03 (permissions).

**DoD Phase 8:**
- [ ] PF-02 редактирование профиля + загрузка аватара
- [ ] NT-01 показывает 3+ типа уведомлений
- [ ] PF-04 переключатели сохраняются
- [ ] PF-08 удаление аккаунта проходит

## Phase 9 — routes (RT-*) + поиск (SR-*)

**Цель**: маршруты в поездке + глобальный поиск.

**DoD Phase 9:**
- [ ] RT-01..04 функциональны
- [ ] SR-01..03 ищут POI/Trips/Entries

## Phase 10 — Состояния + модалки + offline

**Цель**: ST-01..06, MD-01..04 везде работают; черновики синхронизируются после offline.

**DoD Phase 10:**
- [ ] При выключении сети показывается OfflineBanner
- [ ] Запись, созданная offline, отправляется при возобновлении сети
- [ ] Skeletons отображаются при первой загрузке списков

## Phase 11 — Polishing + Observability + CI

**Цель**: production-ready минимум.

Действия:
1. Подключить Prometheus + Grafana (профиль `obs`).
2. Дашборды по основным метрикам.
3. Дописать недостающие тесты до целевого coverage.
4. Schemathesis-фьюзинг в CI.
5. Документация README на каждый сервис.

**DoD Phase 11:**
- [ ] `make cover-check` MIN_COVER=40 зелёный для всех сервисов
- [ ] Dashboards в Grafana показывают rps + p95 + err-rate
- [ ] CI зелёный по всем job из `21_TESTING.md` §4
- [ ] README у каждого сервиса описывает старт + ключевые env

## Phase 12+ — Расширения (post-MVP)

Идеи (не в скоупе MVP):
- Social-вход (Google, Apple)
- Реальная маршрутизация в RT-03 через OSRM
- Реакции/комментарии под записями
- Фолловинг + публичная лента поездок
- Шеринг через `Share sheet`
- Полнотекстовый поиск (Postgres FTS / Meilisearch)
- Конструктор маршрута с оптимизацией порядка точек
- Push на «приближение к POI»

## Связанные файлы

- Все детали реализации каждого пункта → файлы `01..22`
- Изменение этого плана требует обновления `00_INDEX.md`
