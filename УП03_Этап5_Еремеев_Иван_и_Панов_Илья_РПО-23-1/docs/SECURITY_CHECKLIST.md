# SECURITY_CHECKLIST.md — LandMark, этап 5

Проект: LandMark (Wanderlog) · ветка `feat/fifth_stage` · стенд `http://localhost:8080`.
Все проверки выполнены инструментами на развёрнутом demo-стенде; доказательства — в `reports/` и `screenshots/`.

| Проверка | Статус | Доказательство | Что исправлено / примечание |
|---|---|---|---|
| `.env` не загружен в Git | ✅ выполнено | скрин 01, `git ls-files` (трекаются только `*.example`) | `.gitignore`: `.env`, `.env.*`, `**/.env*` |
| `backups`/дампы/логи в `.gitignore` | ✅ выполнено | скрин 01, `.gitignore` | **добавлено**: `backups/`, `*.dump`, `*.bak`, `logs/`, `*.log` |
| `.env.example` без реальных секретов | ✅ выполнено | скрин 02 | `.env.production.example` — только `change_me_*`; demo-значения только в `.env.demo.example` |
| Ручной поиск секретов выполнен | ✅ выполнено | скрин 03, `reports/secret_scan.txt` | реальных паролей/токенов в репозитории не найдено (совпадения — имена колонок в схеме, импорты библиотек, placeholder-ы) |
| Зависимости проверены | ✅ выполнено | скрин 04, `reports/dependency_check.txt` | `govulncheck`: найдены уязвимости; **исправлено** `golang-jwt/jwt/v5` 5.2.1→5.2.2 (GO-2025-3553) |
| Проверены роли (RBAC) | ✅ выполнено | скрин 05, `reports/access_control_check.txt` | anon→401, user→403, admin→200 на `/v1/admin/users`. Gateway срезает клиентский `X-Role` |
| Проверен доступ к чужим данным (IDOR) | ✅ выполнено | скрин 06, `reports/access_control_check.txt` | пользователь B не может читать/менять/удалять поездку пользователя A → 403 |
| CORS/hosts/debug проверены | ✅ выполнено | скрин 07, `reports/cors_before_after.txt` | **исправлено**: `CORS_ORIGINS` `*`→`http://localhost:8080` (конкретный origin) |
| Backup создан | ✅ выполнено | скрин 08, `reports/backup_restore_run.txt` | `pg_dump` → `backups/backup_*.sql` (94 таблицы, 5516 строк) |
| Restore проверен | ✅ выполнено | скрин 09, `reports/backup_restore_run.txt` | восстановление в `landmark_restore_test`, число строк совпало (accounts=12, trips=3, sessions=8) |
| Открытые порты проверены | ✅ выполнено | скрин 10, `reports/ports_check.txt` | БД(5432)/Redis(6379)/бэкенд-сервисы — только внутренняя сеть; наружу только gateway + UI |
| Логи без критических ошибок | ✅ выполнено | скрин 11, `reports/logs_security.txt` | panic/fatal/5xx в сервисах нет; присутствует только объяснимый шум |

**Итог:** 12/12 проверок выполнено. Исправлено 3 проблемы (`.gitignore`, CORS, уязвимая JWT-зависимость).
Остальные замечания вынесены в `RISK_REGISTER.md` и `RECOMMENDATIONS.md`.
