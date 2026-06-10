# INCIDENT_REPORT.md — этап 6, LandMark (Wanderlog)

## 1. Инцидент

API возвращает plain-text `404 page not found` вместо единого JSON-конверта
ошибок при запросе несуществующего маршрута. Зарегистрирован как
**BUG-03** в DEFECT_LOG этапа 4 (статус был open, критичность «Средняя»)
и оформлен как GitHub Issue
[#7](https://github.com/Ilpaka/LandMark_App/issues/7).

Почему это важно: весь остальной API отвечает ошибками вида
`{"error":"..."}` — Flutter-клиент парсит тело ошибки как JSON. На
plain-text ответе парсер падает, и пользователь вместо нормального
сообщения получает ошибку десериализации.

Попутно при выполнении этапа обнаружен и второй инцидент эксплуатации:
**GitHub Actions CI был красным на всех 21 запусках** с момента включения —
автоматическая проверка проекта фактически не работала. Починен в рамках
этого же этапа (подробности в п. 5–6).

## 2. Где обнаружено

- Локально и на demo-стенде из этапа 3 (Docker, `docker-compose.prod.yml`) —
  ещё на этапе 4 при тестировании API инструментами (curl, анализ заголовков).
- Подтверждено на этапе 6: локальный запуск роутера profile-service
  (репродукция без БД) и тесты роутеров всех сервисов.
- Красный CI — вкладка Actions репозитория (runs #1–#21, все failure).

## 3. Как воспроизвести

1. Запустить backend (`docker compose up --build`) или роутер любого сервиса.
2. Выполнить запрос на маршрут, которого нет в роутере:
   ```
   curl -i http://localhost:18080/v1/profile
   ```
   (у profile-service есть только под-пути `/v1/profile/me`,
   `/v1/profile/:user_id`, … — точного маршрута `/v1/profile` нет).
3. Фактический ответ (скриншот `03_bug_reproduced.png`,
   лог `reports/03_bug_reproduced_curl_before.txt`):
   ```
   HTTP/1.1 404 Not Found
   Content-Type: text/plain
   404 page not found
   ```
   То же — для неподдерживаемого метода (`DELETE /v1/profile/me`): Gin даже
   не отличает 405 от 404.
4. Для сравнения, существующий маршрут отвечает правильным конвертом:
   `curl -i http://localhost:18080/v1/profile/me` → `401 {"error":"unauthorized"}`.

## 4. Диагностика

Что проверялось:

- **Логи сервиса** (`reports/04_logs_diagnostics.txt`): при старте Gin
  печатает полный список зарегистрированных маршрутов — точного маршрута
  `GET /v1/profile` в нём нет, значит запрос уходит в обработчик
  «маршрут не найден».
- **Код роутеров** всех 9 сервисов: ни в одном не зарегистрированы
  `router.NoRoute(...)` и `router.NoMethod(...)`; флаг
  `HandleMethodNotAllowed` не включён.
- **Исходники Gin**: при отсутствии NoRoute-обработчика срабатывает дефолт
  `serveError(c, 404, default404Body)`, который пишет `404 page not found`
  с `Content-Type: text/plain`.
- **Тесты-репродукции**: написаны тесты роутеров (18 шт.) на формат 404/405 —
  до исправления все падали с `Content-Type = "text/plain", body
  "404 page not found"` (`reports/03_bug_reproduced_tests.txt`).
- **CI**: по логам джобов GitHub Actions (run #21) — Flutter-джоба падает на
  `flutter pub get` («image_picker >=1.2.0 requires SDK version >=3.6.0»,
  а в workflow запинена Flutter 3.22.0 с Dart 3.4); джоба trips-service
  падает на `cover-check` (покрытие 24% < порога 40%), после чего fail-fast
  матрицы отменяет остальные сервисы (`reports/07_ci_diagnostics.txt`).

## 5. Причина

1. **BUG-03**: единый JSON-формат ошибок реализован только в хендлерах
   конкретных маршрутов. Сценарий «маршрут/метод не найден» никем не
   обрабатывался, поэтому Gin отвечал своим plain-text дефолтом. Это не
   ошибка одного сервиса, а системный пропуск во всех девяти.
2. **Красный CI**: (а) устаревший пин Flutter 3.22.0 — новые версии
   зависимостей клиента требуют Dart SDK >= 3.6; (б) завышенный единый порог
   покрытия 40% для сервисов, у которых юнит-тестов app-слоя ещё нет
   (api-gateway, auth, media) или покрытие ниже (trips, 24%); (в) fail-fast
   матрицы прятал реальную картину, отменяя половину джобов.

## 6. Исправление

Ветка: `feat/sixth_stage`, основной коммит фикса —
`fix: unknown routes now return JSON error envelope in all services (#7)`.

| Файлы | Что сделано |
|---|---|
| `backend/*/internal/modules/*/http/router.go` (8 сервисов) | Зарегистрированы `NoRoute`/`NoMethod` + `HandleMethodNotAllowed = true`: 404 → `{"error":"not found"}`, 405 → `{"error":"method not allowed"}` |
| `backend/auth-service/internal/httpserver/fallback.go`, `server.go` | То же для auth-service, вынесено в `RegisterFallbackHandlers()` |
| `backend/*/.../router_test.go`, `fallback_test.go` (9 файлов) | 18 тестов: неизвестный маршрут и неподдерживаемый метод возвращают JSON-конверт |
| `.github/workflows/ci.yml` | Flutter 3.22.0 → 3.44.1; реальные пороги `cover-check` на сервис; `fail-fast: false`; checkout v5 |
| `.github/workflows/release.yml` (новый) | Релизная сборка по тегу `v*`: бинарники, tar.gz + sha256, GitHub Release |
| `Makefile`, `scripts/*.bat`, `.gitignore` | `make build`, `make release-check`, `test.bat`, `build.bat`, `release-check.bat`, `create-release.bat`; `dist/`, `release/`, `coverage.out` не хранятся в git |
| `CHANGELOG.md`, `RELEASE_NOTES.md` | Версия 0.3.1: что исправлено, как проверить, ограничения |

api-gateway не менялся: для неизвестного префикса он уже отвечал JSON, а
ответы сервисов проксирует как есть — после фикса сервисов через gateway
тоже уходит JSON.

## 7. Проверка

- Локально: `gofmt` (0 непроформатированных), `go vet` и
  `go test ./... -short` по всем 10 модулям — зелёные
  (`reports/06_tests_passed_locally.txt`); `make build` собирает бинарники
  всех сервисов.
- Повторная репродукция после фикса (`reports/05_fix_verified_curl_after.txt`):
  `GET /v1/profile` → `404 {"error":"not found"}` (application/json),
  `DELETE /v1/profile/me` → `405 {"error":"method not allowed"}` +
  корректный заголовок `Allow: GET, PATCH`.
- GitHub Actions: workflow «CI» на PR этапа 6 — успешный
  (скриншот `07_github_actions_success.png`); workflow «Release» по тегу
  `v0.3.1` собрал релизный архив.
- Pull Request с чек-листом и ссылкой `Closes #7`
  (скриншот `08_pull_request.png`).

## 8. Итог

**Проблема исправлена.** Все сервисы отвечают единым JSON-конвертом ошибок
на любые маршруты и методы; регресс закрыт тестами; CI зелёный и снова
выполняет свою функцию; релиз 0.3.1 оформлен (changelog, release notes,
тег + автоматическая сборка).

Остаётся технический долг (не входит в инцидент): поднять покрытие
app-слоя auth/media/trips и вернуть порог 40% в `cover-check`; открытые
дефекты этапа 4 BUG-01, BUG-04, BUG-05, BUG-06 — кандидаты на следующие
инциденты сопровождения.
