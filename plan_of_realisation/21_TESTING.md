# 21 · Стратегия тестирования

> Пирамида: Unit (≥70%) → Integration (~25%) → E2E (~5%). Цели DoD по покрытию фиксируем в `23_ROADMAP.md`.

## 1. Backend

### 1.1 Unit-тесты

- В каждом сервисе: `*_test.go` рядом с тестируемым кодом.
- Покрываем: doменные правила, валидация в use-cases, мапперы, утилиты.
- Запуск: `go test ./... -count=1 -short`.
- Целевое покрытие: ≥ 60% statements в Phase 2.
- В Tennis уже встроен gate `make cover-check` (MIN_COVER=12) — оставляем минимум, в Phase 3 поднимаем.

### 1.2 Integration-тесты

- Каталог `tests/integration/`, build tag `integration`.
- Поднимают **реальные** Postgres + Redis через `ory/dockertest/v3`. Образец — Tennis `auth-service`.
- Тестируют: store-уровень (postgres), Redis-структуры, transactional outbox.
- Запуск: `make test-integration`.

Шаблон:
```go
//go:build integration
package integration

import (
    "context"
    "testing"
    "github.com/ory/dockertest/v3"
    ...
)

func TestStore_PlacesCRUD(t *testing.T) {
    pool := setupPostgres(t)  // testutil
    defer pool.Purge()
    store := postgres.NewStore(pool.DB())
    ...
}
```

### 1.3 E2E (auth)

В `tests/e2e/`, build tag `e2e`. Поднимают весь стек через docker-compose в тестовой сети, прогоняют сценарии:
- регистрация → verify → login → me → refresh → logout
- forgot-password → reset
- admin block → user 401

### 1.4 Contract-фьюзинг (schemathesis)

`scripts/schemathesis.sh` (см. `17_API_CONTRACTS.md`). Запускается отдельной CI-job после поднятия сервиса.

### 1.5 Security checks

- `make sec` — `gosec` static security analysis.
- В Phase 3: `trivy fs` для зависимостей.

## 2. Backend interaction tests (cross-service)

В `tests/interaction/` (новый каталог в `backend/`):
- Поднимаем подмножество сервисов через compose
- Тесты вызывают gateway → проверяют downstream через свои API

Пример: `TestRegistrationCreatesProfile` — auth.register → profile.GET /internal/{user_id} возвращает корректный nickname.

## 3. Flutter

### 3.1 Unit (use-cases, mappers)

`landmark_app/test/`:
- Чистые Dart-тесты без Flutter, использующие `mocktail`.
- Целевое покрытие ≥ 70% domain layer.

### 3.2 Widget-тесты

- Pages-level: `WelcomePage`, `LoginPage`, ListView с empty/loading/error.
- Используем `pumpWidget` с `ProviderScope.overrides=[ ... mock ... ]`.

### 3.3 Golden-тесты (визуальные)

`golden_toolkit`. Снимки для:
- Все основные кнопки и компоненты дизайн-системы (`19_FLUTTER_DESIGN_SYSTEM.md` §6)
- Onboarding 1/2/3
- TR-04, JR-02, MP-04 — три «эталонные» карточки

### 3.4 Integration-тесты

`integration_test/`, исполняются на реальном устройстве/эмуляторе:
- `01_auth_flow_test.dart` — регистрация → verify → main
- `02_create_trip_test.dart` — после логина
- `03_create_entry_test.dart`
- `04_offline_drafts_test.dart` — отключение сети, создание черновика, восстановление, sync

Используем mocked backend (httpbin-like fake) для воспроизводимости.

### 3.5 Static analysis

`analysis_options.yaml`:
```yaml
include: package:flutter_lints/flutter.yaml
analyzer:
  language:
    strict-casts: true
    strict-inference: true
    strict-raw-types: true
linter:
  rules:
    require_trailing_commas: true
    prefer_single_quotes: true
    avoid_dynamic_calls: true
    unawaited_futures: true
```

## 4. CI-матрица

| Job | Триггер | Время | Что проверяет |
|---|---|---|---|
| `backend-unit` | каждый push в `backend/` | ~3 мин | go test -short |
| `backend-integration` | push в main + nightly | ~10 мин | dockertest |
| `backend-e2e` | nightly | ~15 мин | full compose flow |
| `backend-schemathesis` | каждый PR | ~5 мин | OpenAPI compliance |
| `flutter-unit` | каждый push в `landmark_app/` | ~2 мин | flutter test |
| `flutter-golden` | PR | ~3 мин | golden_toolkit |
| `flutter-integration-android` | nightly | ~10 мин | реальный эмулятор |

## 5. Правила тестов

- Никаких `time.Sleep` — всё через `eventually`/`require.Eventually`.
- Ни один тест не должен зависеть от внешней сети (FCM/SMS/SMTP моки).
- Не использовать `httptest.NewServer` для общесервисных flow — только полное `compose up`.

## 6. Связанные файлы

- Auth-тесты, унаследованные → `05_AUTH_SERVICE.md` §5
- CI → `16_DOCKER_INFRA.md` §5
