# landmark_app — клиент Wanderlog (Flutter)

Мобильный клиент приложения **LandMark (Wanderlog)** — журнал путешественника
с картой достопримечательностей. Серверная часть — в каталоге `../backend`.

## Стек

Flutter 3.41 / Dart 3.11 · Material 3 · Riverpod · GoRouter · Dio ·
flutter_map · geolocator · flutter_secure_storage · cached_network_image.

## Архитектура

Clean Architecture, разбиение по фичам:

```
lib/
  bootstrap/   точка входа, DI, конфигурация окружения (env.dart)
  core/        networking, routing, storage, design-system, widgets
  features/    auth, map, places, trips, journal, favorites, notifications,
               profile, settings, search, admin, splash, main_shell
               (каждая фича: data / domain / presentation)
```

## Запуск

```bash
flutter pub get
flutter run
```

Клиент обращается к backend по адресу `http://localhost:8080`
(на Android-эмуляторе — `http://10.0.2.2:8080`). Переопределить можно так:

```bash
flutter run --dart-define=API_BASE_URL=http://<host>:8080
```

> Перед запуском клиента поднимите backend: в корне репозитория `./start-backend.sh`.

## Тесты и анализ

```bash
flutter test       # модульные и виджет-тесты
flutter analyze    # статический анализ
```
