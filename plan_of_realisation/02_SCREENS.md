# 02 · Каталог экранов

> 68 экранов из `WIKI/design/wanderlog_screens.html`. Колонка «Сервис» указывает на бэкенд-источник данных, колонка «Файл» — на feature-модуль во Flutter.
> ID должны 1-в-1 совпадать с дизайном — это позволяет ИИ перепроверять реализацию относительно скриншотов.

## 1. Глобальная навигация

```
Root
├── AuthStack (AU-01..06, SP-01..04)
│   └── после успеха → MainTabs
├── MainTabs (Bottom nav, FAB по центру)
│   ├── Tab 1: Карта     (MP-*, SR-*)
│   ├── Tab 2: Поездки   (TR-*, RT-*)
│   ├── FAB:  Контекстное создание (зависит от вкладки)
│   ├── Tab 3: Журнал    (JR-*)
│   ├── Tab 4: Избранное (FV-*)
│   └── Tab 5: Профиль   (PF-*, NT-*)
└── AdminStack (только role=admin) (AD-01..04)
```

FAB-кнопка контекстная:
- На карте → создание POI (`MP-05`)
- На поездках → создание поездки (`TR-02`)
- На журнале → создание записи (`JR-03`)
- На избранном → ничего (FAB скрыт)
- На профиле → ничего (FAB скрыт)

## 2. Сводная таблица экранов

Колонки:
- **ID** — идентификатор из дизайна
- **Сервис** — основной бэкенд-источник
- **Эндпоинты** — главные REST-вызовы экрана
- **Состояния** — ссылки на `ST-*` (loading/empty/error/offline)
- **Файл Flutter** — путь от `landmark_app/lib/`

| ID | Название | Сервис | Эндпоинты | Состояния | Файл Flutter |
|---|---|---|---|---|---|
| **Splash & Onboarding** ||||||
| SP-01 | Splash | auth | `GET /v1/auth/me` (если refresh) | ST-01 | features/splash/presentation/splash_page.dart |
| SP-02 | Onboarding 1/3 | — | — | — | features/onboarding/.../page_1.dart |
| SP-03 | Onboarding 2/3 | — | — | — | features/onboarding/.../page_2.dart |
| SP-04 | Onboarding 3/3 | — | — | — | features/onboarding/.../page_3.dart |
| **Авторизация** ||||||
| AU-01 | Welcome | — | — | — | features/auth/presentation/welcome_page.dart |
| AU-02 | Регистрация | auth | `POST /v1/auth/register` (новый, см. файл 05) → `POST /v1/auth/verify-email` | ST-02 | features/auth/.../register_page.dart |
| AU-03 | Вход | auth | `POST /v1/auth/login` | ST-02 | features/auth/.../login_page.dart |
| AU-04 | Восстановление | auth | `POST /v1/auth/forgot-password` | ST-02 | features/auth/.../forgot_password_page.dart |
| AU-05 | OTP-код (универс.) | auth | `POST /v1/auth/verify-email` ИЛИ `POST /v1/auth/reset-password` | ST-02 | features/auth/.../otp_page.dart |
| AU-06 | Новый пароль | auth | `POST /v1/auth/reset-password` | ST-02 | features/auth/.../new_password_page.dart |
| **Permissions** ||||||
| PR-01 | Гео pre-prompt | — | system permission_handler | — | features/permissions/.../geo_page.dart |
| PR-02 | Push pre-prompt | notifications | `POST /v1/notifications/devices` (после согласия) | — | features/permissions/.../push_page.dart |
| PR-03 | Камера pre-prompt | — | system | — | features/permissions/.../camera_page.dart |
| **Карта** ||||||
| MP-01 | Карта · главный | places | `GET /v1/places?bbox=...&category=...&limit=200` | ST-01,03 | features/map/.../map_page.dart |
| MP-02 | Фильтры | places | `GET /v1/places/categories` | — | features/map/.../filters_sheet.dart |
| MP-03 | Поиск с подсказками | places + search | `GET /v1/search?q=...&types=place,trip` | — | features/search/.../suggest_page.dart |
| MP-04 | Карточка POI | places + favorites | `GET /v1/places/{id}`, `POST /v1/favorites` | ST-05 | features/map/.../place_detail_page.dart |
| MP-05 | Создание места · 1 | places + media | `POST /v1/places` (draft), `POST /v1/media/uploads` | ST-02 | features/map/.../create_place_step1.dart |
| MP-06 | Создание места · 2 | places | `PATCH /v1/places/{id}` + categories | ST-02 | features/map/.../create_place_step2.dart |
| MP-07 | Отправлено на модерацию | — | — | — | features/map/.../create_place_done.dart |
| MP-08 | Tap по пину · превью | places | `GET /v1/places/{id}` | — | features/map/widgets/pin_preview_sheet.dart |
| **Поездки** ||||||
| TR-01 | Список поездок | trips | `GET /v1/trips?cursor=&limit=20` | ST-04 | features/trips/.../trips_list_page.dart |
| TR-02 | Создание · 1 | trips | `POST /v1/trips` (draft) | — | features/trips/.../create_trip_step1.dart |
| TR-03 | Создание · даты | trips | `PATCH /v1/trips/{id}` | — | features/trips/.../create_trip_dates.dart |
| TR-04 | Карточка поездки | trips + journal | `GET /v1/trips/{id}`, `GET /v1/trips/{id}/entries` | ST-04 | features/trips/.../trip_detail_page.dart |
| TR-05 | Поездка на карте | trips | `GET /v1/trips/{id}/stops` | — | features/trips/.../trip_map_page.dart |
| TR-06 | Поездка на карте (alt) | trips | то же, что TR-05 | — | features/trips/.../trip_map_alt.dart |
| **Журнал** ||||||
| JR-01 | Лента журнала | journal | `GET /v1/journal/entries?cursor=&trip_id=` | ST-04 | features/journal/.../journal_feed_page.dart |
| JR-02 | Запись · просмотр | journal + media | `GET /v1/journal/entries/{id}` | ST-05 | features/journal/.../entry_detail_page.dart |
| JR-03 | Создание записи | journal + media | `POST /v1/journal/entries`, `POST /v1/media/uploads` | ST-02 | features/journal/.../create_entry_page.dart |
| JR-04 | Галерея фото | media | `GET /v1/media/by-entry/{entry_id}` | — | features/journal/.../gallery_page.dart |
| JR-05 | Просмотр фото | media | presigned URL из media | — | features/journal/.../photo_view_page.dart |
| JR-06 | Камера в записи | media | local capture → upload | — | features/journal/.../camera_capture_page.dart |
| **Маршруты** ||||||
| RT-01 | Маршруты в поездке | trips | `GET /v1/trips/{id}/routes` | ST-04 | features/routes/.../routes_list_page.dart |
| RT-02 | Конструктор маршрута | trips | `POST /v1/trips/{id}/routes`, `POST /v1/trips/{id}/routes/{rid}/stops` | — | features/routes/.../builder_page.dart |
| RT-03 | Активная навигация | (offline) | — (статически в MVP) | — | features/routes/.../navigation_page.dart |
| RT-04 | Детали точки | trips | `GET /v1/trips/{id}/routes/{rid}/stops/{sid}` | — | features/routes/.../stop_detail_page.dart |
| **Избранное** ||||||
| FV-01 | Список | favorites | `GET /v1/favorites?type=place` | ST-04 | features/favorites/.../list_page.dart |
| FV-02 | Empty | favorites | — | — | features/favorites/.../empty_view.dart |
| FV-03 | На карте | favorites + places | `GET /v1/favorites?type=place` + bulk POI | — | features/favorites/.../map_view.dart |
| **Профиль и настройки** ||||||
| PF-01 | Профиль | profile + auth | `GET /v1/profile/me`, `GET /v1/auth/me` | ST-01 | features/profile/.../profile_page.dart |
| PF-02 | Редактировать профиль | profile + media | `PATCH /v1/profile/me`, upload avatar | ST-02 | features/profile/.../edit_profile_page.dart |
| PF-03 | Главная настроек | — | — | — | features/profile/.../settings_page.dart |
| PF-04 | Уведомления | notifications | `GET /v1/notifications/preferences`, `PATCH /v1/notifications/preferences` | — | features/profile/.../notifications_settings_page.dart |
| PF-05 | Приватность | profile | `GET /v1/profile/me/privacy`, `PATCH /v1/profile/me/privacy` | — | features/profile/.../privacy_page.dart |
| PF-06 | Синхронизация | (local) | local-only | — | features/profile/.../sync_page.dart |
| PF-07 | Помощь | — | static | — | features/profile/.../help_page.dart |
| PF-08 | Удаление аккаунта | auth | `DELETE /v1/auth/account` | ST-02 | features/profile/.../delete_account_page.dart |
| **Уведомления** ||||||
| NT-01 | Лента | notifications | `GET /v1/notifications?cursor=&limit=` | ST-04 | features/notifications/.../feed_page.dart |
| NT-02 | Empty | notifications | — | — | features/notifications/.../empty_view.dart |
| NT-03 | Push на lock-screen | (system) | — | — | (нативный, конфигурим в platform/) |
| **Поиск** ||||||
| SR-01 | Стартовый | — | (recent + suggest) | — | features/search/.../start_page.dart |
| SR-02 | Результаты | search | `GET /v1/search?q=...&types=place,trip,entry` | ST-04 | features/search/.../results_page.dart |
| SR-03 | Без результатов | — | — | — | features/search/.../empty_view.dart |
| **Админ** ||||||
| AD-01 | Дашборд | moderation | `GET /v1/moderation/stats` | ST-01 | features/admin/.../dashboard_page.dart |
| AD-02 | Очередь | moderation | `GET /v1/moderation/queue?cursor=&limit=` | ST-04 | features/admin/.../queue_page.dart |
| AD-03 | Карточка | moderation | `GET /v1/moderation/items/{id}`, `POST /v1/moderation/items/{id}/decision` | ST-02 | features/admin/.../item_review_page.dart |
| AD-04 | Категории | places | `GET /v1/places/categories`, `POST/PATCH/DELETE /v1/places/categories` | ST-04 | features/admin/.../categories_page.dart |
| **Состояния** ||||||
| ST-01 | Loading skeleton | — | — | — | core/widgets/skeletons/ |
| ST-02 | Что-то сломалось / 500 | — | — | — | core/widgets/error_view.dart |
| ST-03 | Offline | — | connectivity | — | core/widgets/offline_banner.dart |
| ST-04 | Empty (фильтры/без записей) | — | — | — | core/widgets/empty_view.dart |
| ST-05 | 404 (точка не найдена) | — | — | — | core/widgets/not_found_view.dart |
| ST-06 | Toasts/Snackbars | — | — | — | core/utils/snack.dart |
| **Модалки** ||||||
| MD-01 | Confirm dialog | — | — | — | core/widgets/dialogs/confirm.dart |
| MD-02 | Action sheet | — | — | — | core/widgets/dialogs/action_sheet.dart |
| MD-03 | Share sheet | — | system | — | core/widgets/share/share_sheet.dart |
| MD-04 | Date picker (range) | — | — | — | core/widgets/date_range_picker.dart |

## 3. Cross-cutting слой

Эти элементы доступны из любого экрана; задаются один раз в `core/`:

- `OfflineBanner` (висит сверху при потере связи; ST-03)
- `SyncStatusChip` (на ProfilePage, в шапке списка поездок при отложенной отправке)
- `GlobalSnackHost` (для уведомлений ST-06)
- `OnboardingTooltips` (одноразовые подсказки на новых вкладках после первого входа)

## 4. Маппинг ID → URL роутера (GoRouter)

```
/                          → SP-01
/onboarding/1              → SP-02
/onboarding/2              → SP-03
/onboarding/3              → SP-04
/auth/welcome              → AU-01
/auth/register             → AU-02
/auth/login                → AU-03
/auth/forgot               → AU-04
/auth/otp                  → AU-05
/auth/new-password         → AU-06
/permissions/geo|push|camera → PR-01..03
/main/map                  → MP-01
/main/map/filters          → MP-02 (modal)
/main/map/search           → MP-03
/main/places/{id}          → MP-04
/main/places/new/step1     → MP-05
/main/places/new/step2     → MP-06
/main/places/new/done      → MP-07
/main/trips                → TR-01
/main/trips/new/step1      → TR-02
/main/trips/new/dates      → TR-03
/main/trips/{id}           → TR-04
/main/trips/{id}/map       → TR-05
/main/journal              → JR-01
/main/journal/{id}         → JR-02
/main/journal/new          → JR-03
/main/journal/{id}/gallery → JR-04
/main/journal/{id}/photo/{pid} → JR-05
/main/journal/new/camera   → JR-06
/main/trips/{id}/routes    → RT-01
/main/trips/{id}/routes/new → RT-02
/main/trips/{id}/routes/{rid}/active → RT-03
/main/trips/{id}/routes/{rid}/stops/{sid} → RT-04
/main/favorites            → FV-01
/main/favorites/map        → FV-03
/main/profile              → PF-01
/main/profile/edit         → PF-02
/main/profile/settings     → PF-03
/main/profile/settings/notifications → PF-04
/main/profile/settings/privacy → PF-05
/main/profile/settings/sync → PF-06
/main/profile/settings/help → PF-07
/main/profile/settings/delete → PF-08
/main/notifications        → NT-01
/main/search               → SR-01..03
/admin                     → AD-01
/admin/queue               → AD-02
/admin/items/{id}          → AD-03
/admin/categories          → AD-04
```

## 5. Связанные файлы

- Реализация фичей → `20_FLUTTER_FEATURES.md`
- Дизайн-система компонентов → `19_FLUTTER_DESIGN_SYSTEM.md`
- Эндпоинты → `17_API_CONTRACTS.md`
