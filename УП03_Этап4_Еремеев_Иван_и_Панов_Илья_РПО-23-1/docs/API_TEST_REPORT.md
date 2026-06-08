# API_TEST_REPORT.md — LandMark, этап 4

**Базовый URL:** `http://localhost:8080` (API Gateway demo-стенда)
**Инструмент:** `curl` через `tests/api/run_api_tests.sh` (повторяемый прогон).
**Полный лог прогона:** `reports/api_test_run.txt`. **Коллекция Postman:** `reports/postman_collection.json`.

Доступ к API: публичны `GET /v1/places` и `/v1/auth/*`; остальные маршруты требуют `Authorization: Bearer <JWT>`.

## Успешные запросы

| # | Проверка | Метод | Путь | Код | Тело (фрагмент) |
|---|----------|-------|------|-----|-----------------|
| 1 | Health | GET | `/healthz` | **200** | `{"status":"ok"}` |
| 2 | Список мест (публично) | GET | `/v1/places` | **200** | `{"places":[{"title":"Казанский Кремль",...}]}` |
| 3 | Подтверждение email | POST | `/v1/auth/verify-email` | **200** | `{"status":"ok"}` |
| 4 | Вход (получение токена) | POST | `/v1/auth/login` | **200** | `{"access_token":"eyJ..."}` (len 734) |
| 5 | Профиль с токеном | GET | `/v1/profile/me` | **200** | `{"user_id":"...","display_name":"User"}` |
| 6 | Создание поездки | POST | `/v1/trips` | **201** | `{"id":"c89ed9b1-...","title":"QA trip ...","owner_id":"..."}` |

**Проверка БД (TC-05):** после `POST /v1/trips` созданная поездка присутствует в `GET /v1/trips` — действие через API реально изменяет данные в БД. ✔

## Ошибочные сценарии (обработаны корректно, без 500)

| # | Проверка | Метод | Путь | Код | Тело |
|---|----------|-------|------|-----|------|
| 7 | Неверный пароль | POST | `/v1/auth/login` | **401** | `{"error":"invalid_credentials"}` |
| 8 | Невалидный email | POST | `/v1/auth/register` | **400** | `{"error":"bad_request"}` |
| 9 | Короткий пароль | POST | `/v1/auth/register` | **400** | `{"error":"bad_request"}` |
| 10 | Профиль без токена | GET | `/v1/profile` | **401** | `{"error":"unauthorized"}` |
| 11 | Несуществующий UUID места | GET | `/v1/places/000...000` | **404** | `{"error":"not_found"}` |
| 12 | Некорректный UUID места | GET | `/v1/places/not-a-uuid` | **400** | `{"error":"validation_error"}` |

## Итог
- **Пройдено: 12 / 12.** Покрыты коды **200, 201, 400, 401, 404**.
- Ошибочные запросы не приводят к 500 — ошибки обрабатываются и возвращаются в едином JSON-формате `{"error":...}`.
- **Замечание (BUG-03):** запрос несуществующего *маршрута* (например, `GET /v1/profile` без под-пути) возвращает **plain-text** `404 page not found` (дефолт Gin), а не JSON-конверт `{"error":...}` — неоднородность контракта ошибок.
- **Замечание (BUG-02, исправлено):** ранее `POST /v1/auth/register` возвращал `dev_code` (OTP) в теле ответа даже при `APP_ENV=production`. Исправлено — см. `DEFECT_LOG.md` и `reports/defect_devcode_before_after.txt`.
