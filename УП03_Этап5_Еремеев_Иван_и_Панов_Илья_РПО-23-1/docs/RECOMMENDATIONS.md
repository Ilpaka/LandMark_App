# RECOMMENDATIONS.md — LandMark, этап 5

Что сделано на этапе и что рекомендуется сделать дальше для безопасной эксплуатации.

## Уже исправлено на этапе 5
- **`.gitignore`**: добавлены `backups/`, `*.dump`, `*.bak`, `logs/`, `*.log` (дампы и логи не попадут в Git). Миграции `*.sql` сознательно оставлены под контролем версий.
- **CORS**: `Access-Control-Allow-Origin: *` → конкретный origin (`CORS_ORIGINS=http://localhost:8080`; в проде — реальный домен).
- **Зависимость**: `golang-jwt/jwt/v5` 5.2.1 → **5.2.2** (закрыт GO-2025-3553, DoS при разборе JWT-заголовка на gateway). Сборка и стенд обновлены.

## Рекомендуется дальше

### Безопасность
1. **Дообновить зависимости** (из `govulncheck`):
   - `github.com/redis/go-redis/v9` 9.7.0 → **9.7.3** (GO-2025-3540) в `auth-service`;
   - пересобрать образы на **Go 1.25.7+** — закрывает advisories `crypto/tls` (GO-2026-4337 и др.).
   - Проверить Dart-зависимости клиента: `flutter pub outdated` / `dart pub audit` (на этой машине нет Flutter SDK).
2. **Security-заголовки** на gateway: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Strict-Transport-Security`, `Content-Security-Policy`.
3. **HTTPS/TLS**: терминировать на обратном прокси (`deploy/nginx`), редирект 80→443, HSTS.
4. **OTP**: не возвращать `dev_code` в проде (только email/SMS); включать выдачу кода исключительно на demo.
5. **Сменить demo-секреты** перед реальным деплоем: `AUTH_PEPPER`, `INTERNAL_API_KEY`, пароль админа, пароль Grafana (сейчас `admin`).

### Эксплуатация
6. **Регулярные backup**: запускать `backup.bat` по расписанию (Task Scheduler/cron), хранить копии вне сервера, периодически проверять restore.
7. **Операторские UI не наружу**: Grafana/Prometheus/Mailpit/MinIO привязать к `127.0.0.1` или закрыть фаерволом/прокси с авторизацией (сейчас слушают `0.0.0.0`).
8. **Мониторинг**: починить сбор метрик Prometheus (большинство целей DOWN), настроить алерты.
9. **CI security-check**: добавить в пайплайн `govulncheck`, секрет-скан (gitleaks/trufflehog) и проверку, что `.env` не коммитится.
10. **Healthchecks** для Go-сервисов + `depends_on: condition: service_healthy`, чтобы стенд стабильно поднимался после сбоя.
