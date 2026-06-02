# LandMark (Wanderlog) — журнал путешественника с картой достопримечательностей

Мобильное приложение, объединяющее **карту достопримечательностей**, **личный
журнал поездок** и **архив воспоминаний** в едином пространстве, где каждая
запись привязана к месту и хронологии путешествия.

- **Репозиторий:** https://github.com/Ilpaka/LandMark_App
- **Тема (УП.02 → УП.03):** «Мобильное приложение для ведения журнала
  путешественника с картой достопримечательностей»
- **Специальность:** 09.02.07 Информационные системы и программирование
- **ПМ.03:** Сопровождение и обслуживание программного обеспечения

## Что делает программа

- Карта с пинами достопримечательностей (POI), фильтры по категориям, поиск.
- Создание поездок, дневниковых записей с фото, тегами и геометками.
- Маршруты на день, избранное, профиль и настройки приватности.
- Краудсорсинг POI: пользователь предлагает место → очередь модерации → публикация.
- Роли: **гость**, **пользователь**, **администратор** (модерация, категории).

## Технологический стек

| Слой | Технологии |
|---|---|
| Клиент | Flutter 3.41 / Dart 3.11, Material 3, Riverpod, GoRouter, Dio, flutter_map |
| Бэкенд | Go 1.25, Gin, Clean Architecture, 10 микросервисов + API Gateway |
| Данные | PostgreSQL 16 (схема на сервис), Redis 7, S3/MinIO (медиа) |
| Инфраструктура | Docker Compose, Mailpit (почта), Prometheus + Grafana (мониторинг) |
| Авторизация | JWT (RSA, JWKS), OTP по email, роли в claim `role` |

## Структура репозитория

```
LandMark_App/
├── backend/              # Go-микросервисы + docker-compose.yml
│   ├── api-gateway/      # единая точка входа :8080, JWT-проверка, проксирование
│   ├── auth-service/     # регистрация, вход, JWT/JWKS, админ-действия
│   ├── profile-service/  places-service/  trips-service/  journal-service/
│   ├── media-service/    favorites-service/  notifications-service/  moderation-service/
│   ├── pkg/              # общие пакеты (observability)
│   ├── monitoring/       # Prometheus + Grafana
│   └── scripts/init-db.sql
├── landmark_app/         # Flutter-клиент (lib/core + lib/features/*)
├── plan_of_realisation/  # технический план (23 документа, артефакты УП.02)
├── WIKI/                 # отчётность по этапам, дизайн, диаграммы
└── start-backend.sh      # запуск всего бэкенда одной командой
```

## Быстрый запуск

> Подробная пошаговая инструкция — в файле **03_Инструкция_запуска.md**.

```bash
# 1. Backend (нужен запущенный Docker Desktop)
./start-backend.sh                  # либо: cd backend && docker compose up -d --build

# 2. Проверка
curl http://localhost:8080/healthz  # ожидается HTTP 200

# 3. Клиент
cd landmark_app
flutter pub get
flutter run                         # эмулятор Android / симулятор iOS
```

Точки входа после запуска:

| Сервис | Адрес |
|---|---|
| API Gateway | http://localhost:8080 |
| Mailpit (входящие письма) | http://localhost:8025 |
| MinIO Console (медиа) | http://localhost:9001 (minioadmin / minioadmin) |
| Grafana (мониторинг) | http://localhost:3000 (admin / admin) |
| Prometheus | http://localhost:9090 |

Демо-администратор: `admin@apple.com` / `AdminPass123456`.

## Состояние проекта (на момент входного аудита УП.03)

Проект **запускается полностью**: backend поднимается в Docker (16 контейнеров),
сквозной сценарий авторизации работает, Go-тесты (14 пакетов) и Flutter-тесты
(11 тестов) проходят. Подробности — в отчёте `01_Входной_аудит_проекта.docx`.
