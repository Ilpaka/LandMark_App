# Установка и запуск — LandMark (Wanderlog)

Документ описывает запуск проекта через единые команды (BAT / Makefile / Docker).
Проект состоит из двух частей: **backend** (Go-микросервисы в Docker) и
**мобильный клиент** (Flutter).

## Требования

| ПО | Версия | Проверка |
|---|---|---|
| Docker Desktop | 24+ | `docker --version` |
| Flutter SDK | 3.24+ (Dart 3.4+) | `flutter --version` |
| Make (по желанию, для `make ...`) | любой | `make --version` |

> Для backend нужен только Docker — Go устанавливать не требуется, сборка идёт
> внутри контейнеров. Flutter нужен для мобильного клиента.

## Быстрый старт

### Вариант 1. Windows (BAT-скрипты)

```bat
scripts\setup.bat          :: установить зависимости (Flutter + Go)
scripts\docker-up.bat      :: поднять backend в Docker
scripts\run-app.bat        :: запустить мобильный клиент
scripts\check.bat          :: проверить качество (go vet + flutter analyze)
scripts\format.bat         :: отформатировать код
scripts\logs.bat           :: смотреть логи
scripts\docker-down.bat    :: остановить контейнеры
```

### Вариант 2. macOS / Linux (Makefile)

```bash
make help          # список всех команд
make setup         # установить зависимости
make docker-up     # поднять backend (detached)
make run-app       # запустить мобильный клиент (flutter run)
make check         # go vet + flutter analyze + проверка gofmt
make format        # gofmt + dart format
make logs          # логи контейнеров
make docker-down   # остановить контейнеры
```

### Вариант 3. Напрямую через Docker

```bash
docker compose up --build       # поднять весь backend
docker compose ps               # статус сервисов
docker compose logs -f          # логи
docker compose down             # остановить
```

## Адреса после запуска

| Сервис | Адрес | Доступ |
|---|---|---|
| API Gateway (точка входа клиента) | http://localhost:8080 | — |
| Проверка здоровья | http://localhost:8080/healthz | HTTP 200 |
| Публичный список мест (JSON в браузере) | http://localhost:8080/v1/places | — |
| Mailpit (письма: OTP, сброс пароля) | http://localhost:8025 | — |
| MinIO Console (медиа) | http://localhost:9001 | minioadmin / minioadmin |
| Grafana (мониторинг) | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | — |

Демо-доступ в приложении: пользователь `demo@apple.com` / `DemoPass123456`,
администратор `admin@apple.com` / `AdminPass123456`.

## Мобильный клиент

Клиент — это Flutter-приложение (Android/iOS), поэтому он собирается и
запускается штатным инструментом Flutter, а не упаковывается в Docker:

```bash
cd landmark_app
flutter pub get
flutter run                                   # на эмуляторе/устройстве
flutter build apk --debug                     # сборка APK
flutter run --dart-define=API_BASE_URL=http://<host>:8080   # свой адрес backend
```

По умолчанию клиент обращается к `http://localhost:8080`
(на Android-эмуляторе — к `http://10.0.2.2:8080`).

## Настройки (.env)

Все переменные имеют безопасные значения по умолчанию, поэтому файл `.env` для
локального запуска не обязателен. Пример — в [`.env.example`](.env.example):

```bash
cp .env.example backend/.env
```

Настоящий `.env` в репозиторий не коммитится.

## Проверка качества

```bash
make check     # gofmt + go vet (10 сервисов) + flutter analyze
make test      # go test -short (10 сервисов) + flutter test
make format    # gofmt -w + dart format
make lint-strict   # golangci-lint, если установлен (опционально)
```
