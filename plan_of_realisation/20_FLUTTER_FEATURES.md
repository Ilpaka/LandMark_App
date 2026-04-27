# 20 · Flutter: feature-модули

> На каждый из главных экранов из `02_SCREENS.md` — рекомендуемая раскладка `data/domain/presentation`. Подробно показан паттерн на примере `auth` и `map`; остальные кратко.

## 1. Зависимости (`pubspec.yaml`)

```yaml
dependencies:
  flutter: { sdk: flutter }
  cupertino_icons: ^1.0.8
  flutter_riverpod: ^2.6.1
  go_router: ^14.4.1
  dio: ^5.7.0
  freezed_annotation: ^2.4.4
  json_annotation: ^4.9.0
  flutter_secure_storage: ^9.2.2
  drift: ^2.20.0
  path_provider: ^2.1.4
  shared_preferences: ^2.3.0
  cached_network_image: ^3.4.1
  flutter_map: ^7.0.2
  flutter_map_marker_cluster: ^1.3.6
  latlong2: ^0.9.1
  geolocator: ^12.0.0
  permission_handler: ^11.3.1
  image_picker: ^1.1.2
  photo_view: ^0.15.0
  firebase_core: ^3.6.0
  firebase_messaging: ^15.1.3
  flutter_local_notifications: ^17.2.3
  intl: ^0.19.0
  lucide_icons: ^0.279.0
  uuid: ^4.5.1
  collection: ^1.18.0

dev_dependencies:
  flutter_test: { sdk: flutter }
  flutter_lints: ^6.0.0
  build_runner: ^2.4.13
  freezed: ^2.5.7
  json_serializable: ^6.8.0
  drift_dev: ^2.20.0
  mocktail: ^1.0.4
  golden_toolkit: ^0.15.0
  integration_test: { sdk: flutter }
```

## 2. Feature: `features/auth/` (детально)

### 2.1 domain

```dart
// domain/entities/user.dart
@freezed
class User with _$User {
  const factory User({
    required String id,
    required String role,
    required bool emailVerified,
    String? email,
    String? phoneE164,
  }) = _User;
  factory User.fromJson(Map<String, dynamic> j) => _$UserFromJson(j);
}

// domain/entities/tokens.dart
@freezed
class Tokens with _$Tokens {
  const factory Tokens({required String accessToken, required String refreshToken, required DateTime expiresAt}) = _Tokens;
}

// domain/repositories/auth_repository.dart
abstract class AuthRepository {
  Future<User?> tryRestoreSession();
  Future<RegistrationStarted> register({required String name, required String email, required String password});
  Future<void> verifyEmail({required String verificationId, required String code});
  Future<User> login({required String email, required String password});
  Future<void> forgotPassword({required String email});
  Future<void> resetPassword({required String resetId, required String code, required String newPassword});
  Future<void> resendEmailOtp();
  Future<User> me();
  Future<void> signOut();
  Future<void> deleteAccount();
}

class RegistrationStarted { final String verificationId; const RegistrationStarted(this.verificationId); }
```

### 2.2 data

```dart
// data/api/auth_api.dart
class AuthApi {
  final Dio dio;
  AuthApi(this.dio);

  Future<Map<String, dynamic>> register(String name, String email, String password) async {
    final r = await dio.post('/v1/auth/register', data: {'name': name, 'email': email, 'password': password});
    return r.data as Map<String, dynamic>;
  }
  Future<void> verifyEmail(String verificationId, String code) async {
    await dio.post('/v1/auth/verify-email', data: {'verification_id': verificationId, 'code': code});
  }
  Future<Map<String, dynamic>> login(String email, String password, {String? deviceId, String? deviceName, String? platform}) async {
    final r = await dio.post('/v1/auth/login', data: {
      'email': email, 'password': password,
      if (deviceId != null) 'device_id': deviceId,
      if (deviceName != null) 'device_name': deviceName,
      if (platform != null) 'platform': platform,
    });
    return r.data as Map<String, dynamic>;
  }
  Future<Map<String, dynamic>> me() async => (await dio.get('/v1/auth/me')).data as Map<String, dynamic>;
  Future<Map<String, dynamic>> refresh(String refreshToken) async =>
      (await dio.post('/v1/auth/refresh', data: {'refresh_token': refreshToken})).data as Map<String, dynamic>;
  Future<void> logout(String access, String refresh) async =>
      await dio.post('/v1/auth/logout', data: {'access_token': access, 'refresh_token': refresh});
  Future<void> forgot(String email) async => await dio.post('/v1/auth/forgot-password', data: {'email': email});
  Future<void> reset(String resetId, String code, String newPassword) async =>
      await dio.post('/v1/auth/reset-password', data: {'reset_id': resetId, 'code': code, 'new_password': newPassword});
  Future<void> deleteAccount() async => await dio.delete('/v1/auth/account');
}

// data/repositories/auth_repository_impl.dart
class AuthRepositoryImpl implements AuthRepository {
  final AuthApi api;
  final SecureStorage storage;
  AuthRepositoryImpl(this.api, this.storage);
  // implementation, использует API + сохраняет/читает токены
}
```

### 2.3 presentation

#### State + Notifier

```dart
@freezed
class AuthState with _$AuthState {
  const factory AuthState.unknown() = _Unknown;
  const factory AuthState.unauthenticated() = _Unauthed;
  const factory AuthState.authenticated(User user) = _Authed;
}

class AuthNotifier extends StateNotifier<AuthState> {
  final AuthRepository repo;
  AuthNotifier(this.repo) : super(const AuthState.unknown()) { _bootstrap(); }
  Future<void> _bootstrap() async {
    final u = await repo.tryRestoreSession();
    state = u == null ? const AuthState.unauthenticated() : AuthState.authenticated(u);
  }
  Future<void> signIn(String email, String password) async { ... }
  Future<RegistrationStarted> register(...) async { ... }
  Future<void> signOut() async { await repo.signOut(); state = const AuthState.unauthenticated(); }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref.read(authRepositoryProvider));
});
```

#### Pages

- `WelcomePage` (AU-01) — два CTA, навигация на register/login.
- `RegisterPage` (AU-02) — форма с валидацией. На submit: `authNotifier.register()` → push `/auth/otp` с `verification_id` и kind=`registration`.
- `LoginPage` (AU-03) — форма + ссылки.
- `ForgotPasswordPage` (AU-04) — после успеха → `/auth/otp` с kind=`reset`, кнопка resend.
- `OtpPage` (AU-05) — универс. компонент с 6-cell input, авто-focus, paste-handling, таймер 60с до resend.
- `NewPasswordPage` (AU-06) — после OTP, оба поля, индикатор силы, требования (галочки success/grey).

#### Валидаторы

```dart
String? validateEmail(String? v) { ... }
String? validatePassword(String? v) {
  if (v == null || v.length < 10) return 'Минимум 10 символов';
  if (!RegExp(r'[A-Za-z]').hasMatch(v)) return 'Нужна буква';
  if (!RegExp(r'[0-9]').hasMatch(v)) return 'Нужна цифра';
  return null;
}
```

## 3. Feature: `features/map/`

### 3.1 domain

```dart
@freezed
class Place with _$Place {
  const factory Place({
    required String id,
    required String title,
    required double latitude,
    required double longitude,
    String? description,
    String? city,
    String? country,
    String? coverMediaId,
    @Default([]) List<String> categorySlugs,
  }) = _Place;
  factory Place.fromJson(Map<String, dynamic> j) => _$PlaceFromJson(j);
}

@freezed
class BBox with _$BBox {
  const factory BBox({required double south, required double west, required double north, required double east}) = _BBox;
}

abstract class PlacesRepository {
  Future<List<Place>> listInBBox(BBox bbox, {List<String>? categorySlugs, String? q, int limit = 200});
  Future<Place> get(String id);
  Future<List<Category>> categories();
  Future<String> createDraft({required String title, required double lat, required double lng});
  Future<void> patchDraft({required String id, ...});
  Future<void> submit(String id);
}
```

### 3.2 presentation

- `MapPage` (MP-01) — `flutter_map` + `MarkerClusterLayerWidget`. Камера-listener дебаунсит и обновляет `placesInBBoxProvider(bbox, categories)`.
- `MapFiltersSheet` (MP-02) — `showModalBottomSheet`, мультивыбор категорий.
- `MapSearchPage` (MP-03) — autocomplete (использует поиск, см. file 17 §5).
- `PlaceDetailPage` (MP-04) — карточка с фото, описанием, категориями, расстояние, кнопки «Сохранить» (favorites) и «Открыть в карте».
- `PinPreviewSheet` (MP-08) — компактный сниппет от tap по пину.
- `CreatePlaceStep1Page` (MP-05) — выбор точки на карте + название.
- `CreatePlaceStep2Page` (MP-06) — описание + категории + 1-5 фото (выбор из галереи / камера → upload в media → подвязка media_ids).
- `CreatePlaceDonePage` (MP-07) — финальный экран «Отправлено на модерацию».

### 3.3 Provider

```dart
final placesInBBoxProvider = FutureProvider.autoDispose.family<List<Place>, BBoxQuery>((ref, q) {
  return ref.read(placesRepositoryProvider).listInBBox(q.bbox, categorySlugs: q.categories);
});
```

## 4. Feature: `features/trips/`

### Сущности

```dart
@freezed class Trip with _$Trip { ... }       // см. file 08
@freezed class TripStop with _$TripStop { ... }
@freezed class Route with _$Route { ... }
```

### Pages
- `TripsListPage` (TR-01)
- `CreateTripStep1` (TR-02), `CreateTripDates` (TR-03)
- `TripDetailPage` (TR-04) — табы «Лента / Карта / Маршруты»
- `TripMapPage` (TR-05/06) — отображение stops на карте с polyline routes

## 5. Feature: `features/journal/`

- `JournalFeedPage` (JR-01)
- `EntryDetailPage` (JR-02)
- `CreateEntryPage` (JR-03) — `image_picker` или `camera_capture`, грид превью, кнопка «Сохранить»
- `GalleryPage` (JR-04), `PhotoViewPage` (JR-05) на `photo_view`
- `CameraCapturePage` (JR-06) — `camera` пакет (опционально, либо просто `image_picker.fromCamera`)

## 6. Feature: `features/routes/`

- `RoutesListPage` (RT-01)
- `RouteBuilderPage` (RT-02) — drag-reorder списка stops
- `NavigationPage` (RT-03) — статика в MVP (без реальной маршрутизации)
- `StopDetailPage` (RT-04)

## 7. Feature: `features/favorites/`

- `FavoritesListPage` (FV-01)
- `FavoritesEmpty` (FV-02)
- `FavoritesMapPage` (FV-03)

## 8. Feature: `features/profile/`

- `ProfilePage` (PF-01) — аватар, никнейм, статистика поездок
- `EditProfilePage` (PF-02)
- `SettingsPage` (PF-03)
- `NotificationsSettingsPage` (PF-04)
- `PrivacyPage` (PF-05)
- `SyncPage` (PF-06)
- `HelpPage` (PF-07)
- `DeleteAccountPage` (PF-08) — двойное подтверждение, ввод пароля, ConfirmDialog (MD-01)

## 9. Feature: `features/notifications/`

- `NotificationsFeedPage` (NT-01) — pull-to-refresh, mark-read on view
- `NotificationsEmpty` (NT-02)

## 10. Feature: `features/admin/` (только role=admin)

- `AdminDashboardPage` (AD-01) — статистика
- `ModerationQueuePage` (AD-02) — список с FIFO
- `ItemReviewPage` (AD-03) — карточка модерации с двумя кнопками «Одобрить / Отклонить + причина»
- `CategoriesPage` (AD-04) — CRUD категорий

## 11. Связь с offline-режимом

См. `18_FLUTTER_ARCHITECTURE.md` §7. Каждый репозиторий, кроме чисто read-only:
- При `OfflineException` записывает intent в `pending_uploads` или `entries_drafts`.
- `SyncManager` догоняет на восстановлении сети.

## 12. Тесты

- Unit: каждый use-case + reducer/notifier.
- Widget: каждый ключевой Page (использовать `golden_toolkit` для скриншот-тестов под MP-01, TR-04, JR-02).
- Integration: `flutter test integration_test/` — full flow «register → verify → login → me».

## 13. Связанные файлы

- Архитектура → `18`
- Дизайн-токены → `19`
- Каталог экранов и роутов → `02`
- API → `17`
