# RELEASE_NOTES.md — LandMark (Wanderlog)

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
