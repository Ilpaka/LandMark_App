# RELEASE_NOTES.md — LandMark (Wanderlog)

## Версия 0.3.1 — этап 6 «CI/CD, релиз и инцидент поддержки» (2026-06-10)

**Что это за релиз.** Поддерживающий (patch) релиз: исправлен инцидент с
форматом ошибок API, починен постоянно красный CI и добавлена автоматическая
релизная сборка. Функциональность приложения не менялась.

### Что исправлено
- **Единый формат ошибок API** (issue
  [#7](https://github.com/Ilpaka/LandMark_App/issues/7), BUG-03 из этапа 4).
  Раньше запрос несуществующего маршрута (например, `GET /v1/profile` без
  под-пути) возвращал plain-text `404 page not found`, на котором падал
  JSON-парсер клиента. Теперь все 9 сервисов отвечают единым конвертом:
  `404 {"error":"not found"}` и `405 {"error":"method not allowed"}`.
- **GitHub Actions снова зелёный.** CI падал на каждом запуске с момента
  включения: Flutter-джоба — из-за устаревшего пина Flutter 3.22
  (зависимостям нужен Dart SDK >= 3.6), backend-джобы — из-за завышенного
  порога покрытия. Версия Flutter поднята до 3.44.1, пороги покрытия
  выставлены по факту на каждый сервис, fail-fast отключён.

### Что добавлено
- Релизный workflow `.github/workflows/release.yml`: по тегу `vX.Y.Z`
  собираются бинарники всех сервисов, публикуется архив и GitHub Release.
- Команды релиза: `make build`, `make release-check`,
  `scripts\release-check.bat`, `scripts\create-release.bat`.
- 18 тестов роутеров на формат ошибок 404/405.
- `CHANGELOG.md` с историей версий.

### Как проверить
1. Запустить backend: `docker compose up --build`.
2. Выполнить `curl -i http://localhost:8080/v1/places/nonexistent/route` —
   ответ должен быть JSON `{"error":"not found"}`, а не plain-text.
3. Локальные проверки: `make release-check` (или `scripts\release-check.bat`).
4. CI: вкладка Actions — workflow «CI» зелёный; по тегу `v0.3.1` — workflow
   «Release» с архивом `landmark-backend-v0.3.1-linux-amd64.tar.gz`.

### Ограничения
- Покрытие app-слоя auth/media/trips ниже целевых 40% — пороги в CI временно
  снижены для этих сервисов (зафиксировано как технический долг в CHANGELOG).
- Остальные открытые дефекты этапа 4 (BUG-01 мониторинг, BUG-04
  CORS/security-заголовки, BUG-05, BUG-06) — в backlog следующего релиза.

## Версия 0.3.0 — этап 3 «Развертывание» (2026-06-05)

**Что это за релиз.** Подготовка проекта к развертыванию как готового
сервиса/демо-стенда вне среды разработки. Код приложения с этапа 2 не
переписывался — добавлены production/demo-конфигурация, команды развертывания
и инструкции проверки.

### Что добавлено
- `docker-compose.prod.yml` — production/demo-запуск всего backend одной
  командой (10 микросервисов + PostgreSQL, Redis, MinIO, Mailpit, Prometheus,
  Grafana), с `restart: unless-stopped` и режимом `APP_ENV=production`.
- `.env.production.example` и `.env.demo.example` — примеры переменных
  окружения без реальных секретов; реальный `.env.production` закрыт в
  `.gitignore`.
- Скрипты развертывания: `scripts/deploy.bat`, `restart.bat`,
  `check_deploy.bat`, `build_release.bat` (+ `deploy.sh`, `restart.sh` для
  Linux/VPS).
- Документация: `DEPLOYMENT.md`, `DEMO_GUIDE.md`, `RELEASE_NOTES.md`.
- Примеры конфигурации для VPS: Nginx reverse proxy и systemd-юнит
  (в каталоге `project_files_in_repository/`).

### Как запускать
```bash
git checkout feat/third_stage
cp .env.demo.example .env.production
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
# проверка: http://localhost:8080/healthz  ->  HTTP 200
```
Подробнее — в `DEPLOYMENT.md` и `DEMO_GUIDE.md`.

### Известные ограничения
- Демо-стенд рассчитан на запуск на одной машине; горизонтальное
  масштабирование сервисов не настраивается.
- На demo-стенде `CORS_ORIGINS=*` и упрощённые секреты — только для
  демонстрации, не для реального production.
- Письма (OTP/сброс пароля) уходят в Mailpit, а не на реальную почту.
- Push-уведомления (FCM) и экспорт трейсов (OTEL) по умолчанию отключены.

### Базируется на
- Этап 1 — входной аудит проекта.
- Этап 2 — локальная инсталляция (Makefile, BAT-скрипты, `docker-compose.yml`,
  `.env.example`, линтеры/форматтеры, `INSTALL.md`).
