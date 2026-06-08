# TEST_PLAN.md — LandMark, этап 4

## 1. Что проверяется
- **Проект:** LandMark (Wanderlog) — бэкенд из 9 Go-микросервисов + API Gateway, инфраструктура PostgreSQL / Redis / MinIO / Mailpit / Prometheus / Grafana.
- **Версия / ветка:** `feat/fourth_stage`, commit `0ee6039`.
- **Адрес / способ запуска:** локальный demo-стенд из этапа 3
  `docker compose --env-file .env.production -f docker-compose.prod.yml up -d`
  API Gateway: `http://localhost:8080` (health: `/healthz`).
- **Тип проекта:** Fullstack. На этой машине проверяется контур **Backend/API** (Flutter SDK не установлен → web-сборка фронтенда не выполнялась). Web-UI диагностируется на развёрнутых интерфейсах стенда (Grafana).

## 2. Основные сценарии

| ID | Сценарий | Ожидаемый результат | Инструмент | Статус |
|----|----------|---------------------|------------|--------|
| TC-01 | Health-check API Gateway | `200 {"status":"ok"}` | curl | passed |
| TC-02 | Публичное чтение списка мест | `200`, JSON со списком мест | curl | passed |
| TC-03 | Регистрация → подтверждение email → вход | `200` + получен JWT access_token | curl | passed |
| TC-04 | Авторизованный запрос профиля | `GET /v1/profile/me` → `200` с токеном | curl | passed |
| TC-05 | Создание поездки (запись в БД) | `POST /v1/trips` → `201`, объект появляется в `GET /v1/trips` | curl + проверка БД | passed |
| TC-06 | Вход с неверным паролем | `401 {"error":"invalid_credentials"}` | curl | passed |
| TC-07 | Регистрация с невалидным email | `400 {"error":"bad_request"}` | curl | passed |
| TC-08 | Регистрация со слишком коротким паролем | `400` | curl | passed |
| TC-09 | Запрос профиля без токена | `401 {"error":"unauthorized"}` | curl | passed |
| TC-10 | Несуществующий ID места (валидный UUID) | `404 {"error":"not_found"}`, не 500 | curl | passed |
| TC-11 | Некорректный UUID места | `400 {"error":"validation_error"}`, не 500 | curl | passed |
| TC-12 | Автотесты бэкенда | `go test ./...` по 8 сервисам — все пакеты `ok` | go test | passed |
| TC-13 | Производительность ключевых эндпоинтов | средн. время ответа измерено (avg/p95/max) | curl timing / k6 | passed |
| TC-14 | Диагностика web-UI стенда | Lighthouse-отчёт по Grafana (Perf/A11y/BP/SEO) | Lighthouse | passed |
| TC-15 | Логи после сценариев | нет критических 5xx/panic/fatal в сервисах | docker logs | passed |
| TC-16 | Здоровье мониторинга | Prometheus targets: 9 из 11 целей DOWN | Prometheus API | failed → BUG-01 |

## 3. Инструменты
- **DevTools / Network / Console:** проверка через браузер (см. screenshots 02–03); сетевые запросы и статусы дублируются в `reports/api_test_run.txt`.
- **Lighthouse:** `npx lighthouse http://localhost:3000/login` → `reports/lighthouse_report.html` (+ JSON).
- **curl / Postman:** `tests/api/run_api_tests.sh` (12 проверок), коллекция `reports/postman_collection.json`.
- **go test:** `tests`/`scripts/test.bat` → `reports/go_test_report.txt` (14 пакетов `ok`).
- **k6 / нагрузка:** `tests/load/basic_load.js` (k6) и `tests/load/perf_curl.sh` (без k6) → `reports/perf_curl_summary.txt`.
- **Логи:** `docker compose logs` → `reports/logs_tail.txt`.

## 4. Итог
- **Критичные ошибки:** 0 необработанных 5xx в API; все ошибочные сценарии возвращают корректные 4xx.
- **Найденные дефекты:** 6 (см. `DEFECT_LOG.md`), из них **1 исправлен** (BUG-02, утечка OTP-кода) с демонстрацией before/after.
- **Некритичные замечания:** мониторинг не собирает метрики с большинства сервисов (BUG-01), неоднородный формат 404 (BUG-03), `CORS: *` без security-заголовков (BUG-04), шумные error-логи Grafana (BUG-05).
- **Вывод:** развёрнутый бэкенд **работоспособен и готов к дальнейшей эксплуатации в режиме demo**. До продакшена необходимо закрыть BUG-01 (наблюдаемость) и BUG-04 (CORS/заголовки безопасности); BUG-02 уже исправлен.
