# 18 · Flutter: Clean Architecture

> Описывает структуру `landmark_app/lib/`, слои, DI, networking, навигацию, оффлайн.

## 1. Структура каталогов

```
landmark_app/lib/
├── main.dart                       # entry point + bootstrap DI
├── app.dart                        # MaterialApp.router + ThemeData
├── bootstrap/
│   ├── env.dart                    # AppEnv (dev/prod)
│   ├── di.dart                     # ProviderScope overrides + bootstrap
│   └── error_logger.dart           # Crashlytics/sentry adapter (Phase 2)
├── core/
│   ├── design/
│   │   ├── tokens.dart             # Цвета, размеры, радиусы (см. file 19)
│   │   ├── typography.dart
│   │   ├── theme.dart              # ThemeData light (Material 3)
│   │   └── components/             # Кнопки, чипы, карточки и т.д.
│   ├── networking/
│   │   ├── api_client.dart         # Dio + interceptors
│   │   ├── auth_interceptor.dart   # 401 → refresh
│   │   ├── error_mapper.dart       # ErrorBody → AppException
│   │   └── api_endpoints.dart
│   ├── storage/
│   │   ├── secure_storage.dart     # flutter_secure_storage обёртка
│   │   ├── kv_storage.dart         # SharedPreferences обёртка
│   │   └── local_db/               # Drift, оффлайн-кэш
│   │       ├── database.dart
│   │       └── tables/
│   ├── routing/
│   │   ├── router.dart             # GoRouter с всеми путями (см. file 02 §4)
│   │   └── route_guards.dart       # is_authenticated, is_admin
│   ├── widgets/
│   │   ├── empty_view.dart
│   │   ├── error_view.dart
│   │   ├── offline_banner.dart
│   │   ├── skeletons/
│   │   ├── dialogs/
│   │   └── bottom_sheets/
│   ├── utils/
│   │   ├── result.dart             # Either-like Result
│   │   ├── pagination.dart
│   │   └── snack.dart
│   └── permissions/
│       └── permissions_service.dart
├── features/
│   ├── auth/                       # см. file 20
│   ├── splash/
│   ├── onboarding/
│   ├── permissions/
│   ├── map/
│   ├── trips/
│   ├── journal/
│   ├── routes/
│   ├── favorites/
│   ├── profile/
│   ├── notifications/
│   ├── search/
│   └── admin/
├── platform/                       # MethodChannel-обёртки если нужно
└── l10n/
    └── ru.arb
```

## 2. Layered Architecture внутри feature

```
features/<feature>/
├── data/
│   ├── dto/                        # JSON-модели (json_serializable)
│   ├── api/                        # *Api с Dio-вызовами
│   ├── local/                      # Drift DAO + кэш
│   └── repositories/               # Реализация domain.Repository
├── domain/
│   ├── entities/                   # plain Dart классы (freezed)
│   ├── repositories/               # абстракции
│   └── use_cases/                  # бизнес-операции (1 на use case)
└── presentation/
    ├── providers/                  # Riverpod providers (state)
    ├── pages/                      # экраны (см. file 02)
    └── widgets/
```

**Правило зависимостей**: presentation → domain ← data. presentation НЕ ИМПОРТИРУЕТ data напрямую.

## 3. State management: Riverpod

- Глобальное `ProviderScope` в `main.dart`.
- Для feature-specific state — `StateNotifierProvider` с `freezed`-state-классами:
  ```dart
  @freezed
  class TripsListState with _$TripsListState {
    const factory TripsListState.loading() = _Loading;
    const factory TripsListState.data(List<Trip> trips, {required bool hasMore, String? cursor}) = _Data;
    const factory TripsListState.error(String message) = _Error;
  }
  ```
- DI через override в `bootstrap/di.dart`:
  ```dart
  ProviderScope(overrides: [
    apiClientProvider.overrideWithValue(ApiClient(...)),
    secureStorageProvider.overrideWithValue(SecureStorageImpl()),
  ], child: const App())
  ```

## 4. Networking: ApiClient

`core/networking/api_client.dart` — единый Dio с интерсепторами:
- `RequestIdInterceptor` — добавляет `X-Request-Id` (uuid)
- `AuthInterceptor` — добавляет `Authorization: Bearer <access>`. На `401` пытается `POST /v1/auth/refresh`. Если refresh успешен — повторяет запрос. Если нет — вызывает `AuthRepository.signOut()` и роутер уходит на `/auth/welcome`.
- `LoggerInterceptor` — log в dev, off в prod.
- `ErrorMapper` — превращает `DioException` в типизированные `AppException`:
  ```dart
  sealed class AppException implements Exception {
    factory AppException.fromError(Object e, StackTrace? st) { ... }
  }
  class NetworkException extends AppException { final int? code; }
  class UnauthorizedException extends AppException {}
  class ConflictException extends AppException { final String? code; }
  class ValidationException extends AppException { final Map<String, dynamic>? details; }
  class ServerException extends AppException {}
  class OfflineException extends AppException {}
  ```

`baseUrl` берётся из `AppEnv.apiBaseUrl` (dev: `http://10.0.2.2:8080` для Android emu / `http://localhost:8080` для iOS sim).

## 5. Auth flow (Flutter)

- `AuthRepository.signIn(email, password)` → POST `/v1/auth/login` → сохраняет `access_token` и `refresh_token` в `flutter_secure_storage`.
- `AuthRepository.refresh()` — atomic, защищён `Lock` (несколько 401 не вызовут N refresh'ей).
- `AuthRepository.me()` → GET `/v1/auth/me`.
- `authStateProvider` (Riverpod) — broadcasts `AuthState`: `unknown | unauthenticated | authenticated(User user)`.
- Splash (SP-01) ждёт `authStateProvider != unknown` и редиректит.

## 6. Роутинг: GoRouter

`core/routing/router.dart`:
```dart
final router = GoRouter(
  initialLocation: '/',
  redirect: (ctx, state) {
    final auth = ctx.read(authStateProvider);
    final loc = state.matchedLocation;
    final isAuth = auth.isAuthenticated;
    final inAuthFlow = loc.startsWith('/auth') || loc.startsWith('/onboarding') || loc == '/';
    if (!isAuth && !inAuthFlow) return '/auth/welcome';
    if (isAuth && inAuthFlow && loc != '/') return '/main/map';
    if (loc.startsWith('/admin') && auth.user?.role != 'admin') return '/main/map';
    return null;
  },
  routes: [
    GoRoute(path: '/', builder: (_,__) => const SplashPage()),
    // ... все из file 02 §4
    ShellRoute(
      builder: (_, __, child) => MainShell(child: child),
      routes: [
        GoRoute(path: '/main/map', builder: ...),
        // ...
      ],
    ),
    // ...
  ],
);
```

## 7. Локальная БД (Drift)

`core/storage/local_db/`:
- `trips_drafts` — поездки, не отправленные на сервер
- `entries_drafts` — черновики записей журнала
- `pending_uploads` — медиа, которые надо ретрайнуть
- `cached_places` — minimal POI кэш для офлайн-карты

`SyncManager` (singleton-провайдер) при появлении сети:
1. Поднимает черновики
2. Создаёт сущности в правильном порядке (media → entry/place)
3. Удаляет драфт по успеху
4. Показывает Snackbar «Синхронизировано N записей»

## 8. Карта: flutter_map

Используем OSM-тайлы:
- В dev: `https://tile.openstreetmap.org/{z}/{x}/{y}.png`
- В prod: свой кэширующий тайл-прокси (Phase 3)

Кластеризация — `flutter_map_marker_cluster`. Для bbox-запросов: при изменении камеры вызываем `places.list(bbox: ...)` с дебаунсом 500 мс.

## 9. Permissions

`core/permissions/permissions_service.dart`:
- `requestLocation({ background: false })` → PR-01
- `requestNotifications()` → PR-02 (на iOS отдельный API; на Android 13+ — POST_NOTIFICATIONS)
- `requestCamera()` → PR-03
- `requestPhotoLibrary()` → системный для iOS

При первом отказе показываем экран PR-* как pre-prompt; при повторном отказе ведём в системные настройки.

## 10. Локализация

`flutter intl` с одним `ru.arb`. Структура: `auth_login_title`, `map_filter_all`, `error_unknown`, …

## 11. Обработка ошибок: единый паттерн

```dart
final result = await asyncResult(() => useCase.call());
result.when(
  data: (value) { ... },
  error: (err) {
    if (err is OfflineException) {
      ref.read(snackProvider).show('Нет соединения');
    } else if (err is ValidationException) {
      ref.read(snackProvider).show(err.firstFieldMessage);
    } else {
      ref.read(snackProvider).show('Что-то пошло не так');
    }
  },
);
```

## 12. Связанные файлы

- Дизайн-система → `19_FLUTTER_DESIGN_SYSTEM.md`
- Feature-модули → `20_FLUTTER_FEATURES.md`
- API-контракты, на которые завязана сеть → `17_API_CONTRACTS.md`
