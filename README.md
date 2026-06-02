# LandMark (Wanderlog)

Мобильное приложение — **журнал путешественника с картой достопримечательностей**:
карта POI, личный журнал поездок с фото и геометками, маршруты, избранное и
краудсорсинг мест с модерацией.

| | |
|---|---|
| **Клиент** | Flutter 3.41 / Dart 3.11 · Material 3 · Riverpod · GoRouter · Dio · flutter_map |
| **Бэкенд** | Go 1.25 + Gin · Clean Architecture · 10 микросервисов + API Gateway |
| **Данные** | PostgreSQL 16 (схема на сервис) · Redis 7 · S3/MinIO |
| **Инфраструктура** | Docker Compose · Mailpit · Prometheus + Grafana |

## Структура репозитория

```
backend/              Go-микросервисы + docker-compose.yml
  api-gateway/        единая точка входа :8080, JWT-проверка, проксирование
  auth-service/       регистрация, вход, JWT/JWKS, админ-действия
  profile/ places/ trips/ journal/ media/ favorites/ notifications/ moderation/
  pkg/                общие пакеты (observability)
  monitoring/         Prometheus + Grafana
landmark_app/         Flutter-клиент (lib/core + lib/features/*)
plan_of_realisation/  технический план (23 документа)
WIKI/                 отчётность по этапам, дизайн, диаграммы
start-backend.sh      запуск всего бэкенда одной командой
```

## Быстрый запуск

```bash
# 1. Backend (нужен запущенный Docker Desktop)
./start-backend.sh                       # = cd backend && docker compose up -d --build

# 2. Проверка
curl http://localhost:8080/healthz       # ожидается HTTP 200

# 3. Клиент
cd landmark_app && flutter pub get && flutter run
```

Точки входа после запуска:

| Сервис | Адрес | Доступ |
|---|---|---|
| API Gateway | http://localhost:8080 | — |
| Mailpit (письма) | http://localhost:8025 | — |
| MinIO Console | http://localhost:9001 | minioadmin / minioadmin |
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | — |

Демо-администратор: `admin@apple.com` / `AdminPass123456`.

Пример переменных окружения — в [`backend/.env.example`](backend/.env.example)
(все значения по умолчанию безопасны, файл `.env` для локального запуска не обязателен).

## Тесты

```bash
# backend (по каждому сервису)
cd backend/auth-service && go test ./... -short
# клиент
cd landmark_app && flutter test
```

## Статус

Проект запускается полностью (backend — 16 контейнеров в Docker), serverside-тесты
(14 пакетов) и Flutter-тесты (11 тестов) проходят. Подробный входной аудит —
в каталоге `УП03_Этап1_*` (отчёт `01_Входной_аудит_проекта.docx`).
