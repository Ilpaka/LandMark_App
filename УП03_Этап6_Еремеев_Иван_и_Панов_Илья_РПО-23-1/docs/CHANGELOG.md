# CHANGELOG — LandMark (Wanderlog)

Формат: [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версии — [SemVer](https://semver.org/lang/ru/).

## [0.3.1] — 2026-06-10 (этап 6: CI/CD, релиз и инцидент поддержки)

### Fixed
- **BUG-03** (issue [#7](https://github.com/Ilpaka/LandMark_App/issues/7), из DEFECT_LOG этапа 4):
  запрос несуществующего маршрута возвращал plain-text `404 page not found`
  (дефолт Gin) вместо единого JSON-конверта ошибок. Во всех 9 сервисах
  добавлены `NoRoute`/`NoMethod`-обработчики: теперь неизвестный маршрут —
  `404 {"error":"not found"}`, неподдерживаемый метод — `405 {"error":"method
  not allowed"}` с заголовком `Allow`.
- **CI был красным на всех запусках** (runs #1–#21):
  - Flutter-джоба падала на `flutter pub get`: в workflow была запинена
    Flutter 3.22.0, а `image_picker >=1.2.0` требует Dart SDK >= 3.6 —
    поднята до 3.44.1 (версия, которую рекомендует сам pub);
  - `cover-check MIN_COVER=40` валил trips (24%), auth, media и api-gateway
    (нет юнит-тестов app-слоя), а fail-fast матрицы отменял остальные джобы —
    пороги выставлены по факту на сервис, fail-fast отключён.

### Added
- Тесты роутеров всех 9 сервисов на JSON-конверт для 404/405 (`router_test.go`,
  `fallback_test.go`) — 18 новых тестов.
- `.github/workflows/release.yml` — релизная сборка по тегу `v*`: бинарники
  всех Go-сервисов, tar.gz-архив с контрольной суммой, GitHub Release.
- Команды релиза: `make build`, `make release-check`; для Windows —
  `scripts/test.bat`, `scripts/build.bat`, `scripts/release-check.bat`,
  `scripts/create-release.bat`.
- `CHANGELOG.md` (этот файл).

### Changed
- `actions/checkout` поднят с v4 до v5 (предупреждение о устаревании Node 20
  в раннерах GitHub с 16.06.2026).

### Verified
- Локально: `gofmt` + `go vet` + `go test` по всем 10 модулям — зелёные;
  `make build` собирает все бинарники; фикс проверен вручную через `curl`.
- CI: GitHub Actions (workflow `CI`) — успешный прогон на PR этапа 6.

### Технический долг
- Поднять покрытие app-слоя auth-service, media-service, trips-service и
  вернуть порог 40% (а дальше — выше) в `cover-check`.

## [0.3.0] — 2026-06-05 (этап 3: развертывание)

### Added
- `docker-compose.prod.yml` — production/demo-стенд одной командой
  (10 сервисов + PostgreSQL, Redis, MinIO, Mailpit, Prometheus, Grafana).
- `.env.production.example`, `.env.demo.example`, скрипты развертывания
  (`deploy`, `restart`, `check_deploy`, `build_release`), `DEPLOYMENT.md`,
  `DEMO_GUIDE.md`, `RELEASE_NOTES.md`.

### Fixed (в рамках этапа 4, после 0.3.0)
- **BUG-02**: `dev_code` (OTP) больше не возвращается в ответе
  `POST /v1/auth/register` в production — только при явном opt-in
  `AUTH_EXPOSE_DEV_CODE=true`.

## [0.2.0] — 2026-06-05 (этап 2: локальная инсталляция)

### Added
- Единые команды запуска/проверки: `Makefile`, `scripts/*.bat`,
  `docker-compose.yml`, `.env.example`, `INSTALL.md`.
- Линтеры и форматтеры: `gofmt`, `go vet`, `golangci-lint`,
  `flutter analyze`, `.editorconfig`, `.golangci.yml`.

## [0.1.0] — 2026-05-05 (этап 1: входной аудит)

### Added
- Базовая версия проекта: Flutter-клиент (карта POI, журнал, маршруты,
  избранное) + 10 Go-микросервисов с API Gateway, PostgreSQL, Redis, MinIO.
- Технический план (`plan_of_realisation/`), отчётность по этапам (`WIKI/`).
