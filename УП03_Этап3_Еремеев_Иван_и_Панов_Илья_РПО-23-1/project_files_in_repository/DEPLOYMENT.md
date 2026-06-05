# DEPLOYMENT.md — развертывание LandMark (Wanderlog)

Документ описывает, как развернуть проект **как готовый сервис (демо-стенд)**,
проверить его доступность, посмотреть логи и перезапустить — вне среды разработки.

Проект состоит из двух частей:

- **backend** — 10 Go-микросервисов + инфраструктура (PostgreSQL, Redis, MinIO,
  Mailpit, Prometheus, Grafana), запускается в Docker;
- **мобильный/web-клиент** на Flutter — собирается отдельным release-билдом
  (см. [DEMO_GUIDE.md](DEMO_GUIDE.md) и `scripts/build_release.bat`).

---

## 1. Где развернут проект

**Вариант развертывания:** C — локальный demo-стенд через Docker
(production-команда `docker compose -f docker-compose.prod.yml up -d`).

Проект представляет собой систему из 10 микросервисов с базой данных и
требует приватных ключей/секретов, поэтому для учебной демонстрации он
поднимается как **production-демо-стенд** одной командой. Тот же
`docker-compose.prod.yml` без изменений разворачивается на VPS/Linux —
отличаются только адрес и значения `.env.production`.

**Адрес после запуска:**

| Сервис | Адрес | Назначение |
|--------|-------|------------|
| API Gateway | http://localhost:8080 | Точка входа REST API (health: `/healthz`) |
| Grafana | http://localhost:3000 | Мониторинг (логин `admin`) |
| Prometheus | http://localhost:9090 | Сбор метрик |
| Mailpit | http://localhost:8025 | Входящие письма (OTP, сброс пароля) |
| MinIO Console | http://localhost:9001 | Хранилище медиа |

> На VPS вместо `localhost` подставляется домен/IP сервера, а наружу
> рекомендуется открывать только порт API Gateway (через Nginx reverse proxy).

---

## 2. Требования

- **ОС:** Windows 10/11, Linux или macOS.
- **Docker:** Docker Desktop / Docker Engine **20.10+** и Docker Compose **v2**.
- **Порты:** 8080 (API), 3000 (Grafana), 9090 (Prometheus), 8025 (Mailpit),
  9000/9001 (MinIO). Должны быть свободны.
- **База данных:** PostgreSQL 16 — поднимается автоматически как контейнер
  (отдельно ставить не нужно).
- **Переменные окружения:** файл `.env.production` (создаётся из
  `.env.production.example` или `.env.demo.example`).
- Для сборки клиента (необязательно): Flutter SDK 3.x.

---

## 3. Команды развертывания

```bash
# 1. Клонировать репозиторий и перейти в ветку этапа 3
git clone https://github.com/Ilpaka/LandMark_App
cd LandMark_App
git checkout feat/third_stage

# 2. Подготовить переменные окружения (секреты НЕ коммитятся)
cp .env.demo.example .env.production        # для демо-стенда
# или: cp .env.production.example .env.production  # и заполнить реальные секреты

# 3. Собрать и запустить весь стенд в фоне
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d

# 4. Проверить статус
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

На Windows те же действия выполняются одной командой: **`scripts\deploy.bat`**
(скрипт сам создаёт `.env.production` из примера, если его нет).

---

## 4. Проверка

- **Адрес приложения:** http://localhost:8080
- **Проверка работоспособности (health-check):**
  ```bash
  curl -i http://localhost:8080/healthz       # ожидается HTTP 200
  ```
- **Команда просмотра логов:**
  ```bash
  docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=80
  ```
- **Проверка статуса контейнеров:** `docker compose ... ps` — все сервисы в
  состоянии `running` / `healthy`.
- **Основной сценарий проверки:** регистрация пользователя → подтверждение OTP
  (письмо видно в Mailpit http://localhost:8025) → вход → запрос списка мест
  через API Gateway. Подробный сценарий — в [DEMO_GUIDE.md](DEMO_GUIDE.md).

Готовая проверка одной командой (Windows): **`scripts\check_deploy.bat`** —
печатает статус, код ответа `/healthz` и хвост логов.

---

## 5. Остановка и перезапуск

```bash
# Остановить стенд
docker compose --env-file .env.production -f docker-compose.prod.yml down

# Запустить заново
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
```

На Windows — **`scripts\restart.bat`** (на Linux/VPS — `scripts/restart.sh`).
Данные БД и медиа сохраняются между перезапусками в Docker volumes
(`pgdata`, `miniodata`).

---

## 6. Развертывание на VPS/Linux (опционально)

Тот же `docker-compose.prod.yml` работает на VPS. Кратко:

```bash
ssh user@server_ip
git clone https://github.com/Ilpaka/LandMark_App && cd LandMark_App
git checkout feat/third_stage
cp .env.production.example .env.production   # заполнить реальные секреты
bash scripts/deploy.sh
docker ps                                    # убедиться, что сервисы running
```

Для публичного доступа перед API Gateway ставится Nginx reverse proxy
(пример конфигурации — `deploy/nginx/landmark.conf`), а автозапуск стенда
обеспечивает systemd-юнит (`deploy/systemd/landmark.service`).

---

## 7. Частые проблемы

| Симптом | Причина | Решение |
|---------|---------|---------|
| `port is already allocated` | Порт 8080/3000/9000 занят | Освободить порт или изменить маппинг в compose |
| API отвечает 502/не отвечает | Сервисы ещё собираются/стартуют | Подождать сборку, проверить `... logs` |
| OTP-письмо «не приходит» | Письма уходят в Mailpit, не на реальную почту | Открыть http://localhost:8025 |
| `.env.production` с реальными секретами в git | Файл не должен коммититься | Он закрыт в `.gitignore`, в репозитории только `*.example` |
