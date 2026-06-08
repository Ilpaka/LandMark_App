# DEFECT_LOG.md — LandMark, этап 4

Все дефекты найдены инструментами на развёрнутом demo-стенде (curl, Prometheus API,
анализ заголовков ответа, чтение исходного кода и логов). Дефект — это любая выявленная
проблема, а не только полный отказ.

| ID | Где найдено | Описание проблемы | Как воспроизвести | Критичность | Статус | Исправление |
|----|-------------|-------------------|-------------------|-------------|--------|-------------|
| BUG-01 | Мониторинг / Prometheus | Prometheus не может собрать метрики с **9 из 11** целей: `api-gateway` отдаёт `/metrics` с **401**, 8 сервисов — **404**. UP только `auth-service` и сам `prometheus`. Мониторинг фактически «слепой». | `curl http://localhost:9090/api/v1/targets` → у большинства `health=down`; или Prometheus UI → Status → Targets | Высокая | **open** | Открыть `/metrics` без JWT (gateway) и добавить метрик-эндпоинт сервисам; либо скрейпить сервисы напрямую во внутренней сети |
| BUG-02 | API / auth-service | `POST /v1/auth/register` возвращал OTP-код подтверждения (`dev_code`) в теле HTTP-ответа **без проверки окружения** — даже при `APP_ENV=production`. Утечка одноразового кода = риск захвата аккаунта. | `curl -X POST /v1/auth/register -d '{...}'` → в ответе поле `"dev_code"` | Высокая | **fixed** | Код gate-нут `shouldExposeDevCode()` (флаг `AUTH_EXPOSE_DEV_CODE`, безопасно по умолчанию) + проброс переменной в `docker-compose.prod.yml`. Демонстрация before/after: `reports/defect_devcode_before_after.txt`, скрин `09_defect_before_after.png` |
| BUG-03 | API / контракт ошибок | Запрос несуществующего **маршрута** (например `GET /v1/profile` без под-пути) возвращает **plain-text** `404 page not found` (дефолт Gin) вместо единого JSON-конверта `{"error":"not_found"}`, который отдают остальные ошибки. Клиент, парсящий JSON, упадёт. | `curl -i http://localhost:8080/v1/profile` (с валидным токеном) → `Content-Type: text/plain`, тело `404 page not found` | Средняя | **open** | Добавить `router.NoRoute(...)` с JSON-ответом во всех сервисах (или централизованно на gateway) |
| BUG-04 | API / безопасность | На ответах API стоит **`Access-Control-Allow-Origin: *`** (разрешён любой источник) и **отсутствуют** security-заголовки: нет `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`, `Content-Security-Policy`. | `curl -I http://localhost:8080/v1/places` → виден `Access-Control-Allow-Origin: *`, заголовков безопасности нет | Средняя | **open** | Ограничить CORS списком доверенных доменов (`CORS_ORIGINS`), добавить security-заголовки в middleware gateway |
| BUG-05 | Логи / Grafana | В логах Grafana при каждом старте повторяются `level=error` о ненайденных каталогах provisioning (`plugins`, `notifiers`, `alerting`). Не влияет на работу, но засоряет логи и мешает диагностике. | `docker compose logs grafana \| grep level=error` (см. `reports/logs_tail.txt`) | Низкая | **open** | Смонтировать пустые каталоги provisioning или отключить их в конфиге Grafana |
| BUG-06 | Конфигурация / Go-сервисы | Сервисы запущены в **debug-режиме Gin** при `APP_ENV=production`: в логах `[WARNING] Running in "debug" mode. Switch to "release" mode in production.` Лишнее подробное логирование и накладные расходы; в логи попадает полный список маршрутов. | `docker compose logs trips-service \| grep -i "debug.*mode"` (видно и на скрине `08_logs_without_critical_errors.png`) | Низкая | **open** | Выставить `GIN_MODE=release` (env) для Go-сервисов в production-конфиге |

## Сводка
- Всего дефектов: **6**. Исправлено: **1** (BUG-02). Открыто: 5 (Низкая — 3, Средняя — 2, Высокая — 1).
- **Критических необработанных ошибок в API нет:** все ошибочные сценарии возвращают корректные коды 4xx, 5xx/panic в логах сервисов отсутствуют.
- Приоритет на доработку до продакшена: **BUG-01** (наблюдаемость) и **BUG-04** (CORS/заголовки безопасности).

## Подробно: BUG-02 (исправлено) — before / after
**Before:** `handlers.go` отдавал `dev_code` безусловно:
```go
c.JSON(http.StatusOK, gin.H{"status": "ok", "verification_id": verID.String(), "dev_code": plainCode})
```
**After:** код возвращается только при явном opt-in (demo), production безопасен по умолчанию:
```go
resp := gin.H{"status": "ok", "verification_id": verID.String()}
if h.ExposeDevCode { // AUTH_EXPOSE_DEV_CODE=true или APP_ENV != production
    resp["dev_code"] = plainCode
}
c.JSON(http.StatusOK, resp)
```
Проверка на стенде (`reports/defect_devcode_before_after.txt`):
- production-safe (`AUTH_EXPOSE_DEV_CODE=false`) → `{"status":"ok","verification_id":"..."}` — кода нет;
- demo opt-in (`=true`) → `{"dev_code":"211139","status":"ok",...}` — demo-сценарий работает.
