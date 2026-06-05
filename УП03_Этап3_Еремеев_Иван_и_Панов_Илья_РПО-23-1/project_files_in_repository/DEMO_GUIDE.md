# DEMO_GUIDE.md — как быстро проверить LandMark (Wanderlog)

Документ для преподавателя/проверяющего: как за несколько минут поднять стенд
и убедиться, что проект работает.

## Как быстро проверить проект

1. Запустить демо-стенд (backend в Docker):
   - **Windows:** `scripts\deploy.bat`
   - **Linux/macOS:** `bash scripts/deploy.sh`
   - вручную:
     ```bash
     cp .env.demo.example .env.production
     docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
     ```
2. Дождаться, пока все контейнеры перейдут в состояние `running`/`healthy`
   (проверка: `scripts\check_deploy.bat`).
3. Проверить основной сценарий (через браузер/`curl`):
   - **действие 1 — health-check:** открыть http://localhost:8080/healthz →
     ожидается ответ `HTTP 200`.
   - **действие 2 — регистрация и OTP:** через API Gateway зарегистрировать
     пользователя; письмо с кодом подтверждения появится в Mailpit
     http://localhost:8025.
   - **действие 3 — вход и данные:** войти под demo-аккаунтом и запросить
     список мест/профиль через API Gateway (REST `/api/...`).
   - **действие 4 — мониторинг:** открыть Grafana http://localhost:3000
     (логин `admin`) и увидеть метрики сервисов.

Мобильный/web-клиент собирается отдельно командой `scripts\build_release.bat`
(Flutter web) и подключается к backend по адресу `API_BASE_URL`.

## Тестовые данные (demo-стенд)

```
Администратор:
  Логин:  admin@apple.com
  Пароль: AdminPass123456

Grafana:
  Логин:  admin
  Пароль: admin   (значение GRAFANA_PASSWORD из .env.demo.example)
```

> Это demo-значения из `.env.demo.example`. В реальном production они задаются
> в `.env.production` и в репозиторий не попадают.

## Что считается успешной проверкой

- приложение/стенд открылось (API Gateway отвечает `200` на `/healthz`);
- основной сценарий выполнен (регистрация/вход/запрос данных проходят);
- в логах нет критических ошибок (`scripts\check_deploy.bat`);
- стенд останавливается и снова запускается (`scripts\restart.bat`).
