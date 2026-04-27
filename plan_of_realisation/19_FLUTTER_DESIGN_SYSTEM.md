# 19 · Flutter: дизайн-система

> Карта дизайна `WIKI/design/wanderlog_screens.html` → конкретные Dart-токены и компоненты. Цель — в пиксель повторить дизайн.

## 1. Цвета (от 01.01 — Палитра)

```dart
// core/design/tokens.dart
abstract final class AppColors {
  // Brand · Teal
  static const teal900 = Color(0xFF063737);
  static const teal700 = Color(0xFF0E5C5B);
  static const teal600 = Color(0xFF0E7C7B); // primary
  static const teal500 = Color(0xFF159594);
  static const teal300 = Color(0xFF7BC9C8);
  static const teal100 = Color(0xFFD6EFEE);
  static const teal50  = Color(0xFFEAF7F6);

  // Accent · Coral
  static const coral700 = Color(0xFFC84A2D);
  static const coral600 = Color(0xFFE85B3A);
  static const coral500 = Color(0xFFFF6B47); // accent
  static const coral300 = Color(0xFFFFB099);
  static const coral100 = Color(0xFFFFE2D9);
  static const coral50  = Color(0xFFFFF1EB);

  // Sand
  static const sand700 = Color(0xFF8A7457);
  static const sand500 = Color(0xFFC9B392);
  static const sand300 = Color(0xFFE5D7BC);
  static const sand100 = Color(0xFFF5EBDD);
  static const sand50  = Color(0xFFFBF5EC);

  // Neutrals
  static const cream    = Color(0xFFFFF8EE);
  static const paper    = Color(0xFFFFFDF8);
  static const ink      = Color(0xFF1A2026);
  static const inkSoft  = Color(0xFF2D3438);
  static const slate700 = Color(0xFF3A434A);
  static const slate500 = Color(0xFF6B7280);
  static const slate400 = Color(0xFF9AA3AD);
  static const slate300 = Color(0xFFC7CDD3);
  static const slate200 = Color(0xFFE5E8EB);
  static const slate100 = Color(0xFFF1F3F5);
  static const white    = Color(0xFFFFFFFF);

  // Semantic
  static const success      = Color(0xFF0F9D7B);
  static const successLight = Color(0xFFD8F4E9);
  static const warn         = Color(0xFFE7A400);
  static const warnLight    = Color(0xFFFFF3D0);
  static const error        = Color(0xFFDC4A3D);
  static const errorLight   = Color(0xFFFBE0DC);
  static const info         = Color(0xFF2E7BD7);
  static const infoLight    = Color(0xFFDFEBF9);
}
```

## 2. Типографика (от 01.02)

Шрифты: **Fraunces** (display, serif) и **Manrope** (body). `JetBrains Mono` для технических меток. Скачать через `google_fonts: ^6.2` либо положить файлы в `assets/fonts/`. Рекомендуется: **локальные файлы** (стабильнее, без зависимости от сети).

`pubspec.yaml`:
```yaml
flutter:
  fonts:
    - family: Fraunces
      fonts:
        - asset: assets/fonts/Fraunces-Variable.ttf
        - asset: assets/fonts/Fraunces-Italic-Variable.ttf
          style: italic
    - family: Manrope
      fonts:
        - asset: assets/fonts/Manrope-Variable.ttf
    - family: JetBrainsMono
      fonts:
        - asset: assets/fonts/JetBrainsMono-Variable.ttf
```

```dart
// core/design/typography.dart
abstract final class AppTypography {
  static const display = TextStyle(fontFamily: 'Fraunces', fontWeight: FontWeight.w400, fontSize: 96, height: 0.98, letterSpacing: -1.92, color: AppColors.ink);
  static const h1      = TextStyle(fontFamily: 'Fraunces', fontWeight: FontWeight.w500, fontSize: 32, letterSpacing: -0.32, color: AppColors.ink);
  static const h2      = TextStyle(fontFamily: 'Fraunces', fontWeight: FontWeight.w500, fontSize: 24, color: AppColors.ink);
  static const bodyL   = TextStyle(fontFamily: 'Manrope',  fontWeight: FontWeight.w500, fontSize: 16, color: AppColors.ink);
  static const body    = TextStyle(fontFamily: 'Manrope',  fontWeight: FontWeight.w400, fontSize: 14, color: AppColors.ink);
  static const caption = TextStyle(fontFamily: 'Manrope',  fontWeight: FontWeight.w600, fontSize: 12, letterSpacing: 0.24, color: AppColors.slate500);
  static const mono    = TextStyle(fontFamily: 'JetBrainsMono', fontSize: 11, letterSpacing: 0.88, color: AppColors.coral600);

  static const buttonL = TextStyle(fontFamily: 'Manrope', fontWeight: FontWeight.w600, fontSize: 16, color: AppColors.white);
  static const button  = TextStyle(fontFamily: 'Manrope', fontWeight: FontWeight.w600, fontSize: 15);
}
```

## 3. Радиусы и тени (01.03 — 01.04)

```dart
abstract final class AppRadius {
  static const r8  = BorderRadius.all(Radius.circular(8));   // chips, tags
  static const r14 = BorderRadius.all(Radius.circular(14));  // buttons, inputs
  static const r16 = BorderRadius.all(Radius.circular(16));  // cards
  static const r20 = BorderRadius.all(Radius.circular(20));  // big cards
  static const r28 = BorderRadius.all(Radius.circular(28));  // bottom sheets
}
abstract final class AppShadows {
  static const sm = [BoxShadow(color: Color(0x10141E28), offset: Offset(0,1), blurRadius: 2)];
  static const md = [BoxShadow(color: Color(0x14141E28), offset: Offset(0,4), blurRadius: 18)];
  static const lg = [BoxShadow(color: Color(0x1F141E28), offset: Offset(0,18), blurRadius: 60)];
  static const fab = [BoxShadow(color: Color(0x66FF6B47), offset: Offset(0,8), blurRadius: 18)];
}
```

## 4. Spacing

```dart
abstract final class AppSpacing {
  static const xs  = 4.0;
  static const sm  = 8.0;
  static const md  = 12.0;
  static const lg  = 16.0;
  static const xl  = 24.0;
  static const xxl = 32.0;
}
```

## 5. ThemeData (Material 3)

```dart
// core/design/theme.dart
ThemeData buildLightTheme() {
  final scheme = ColorScheme.fromSeed(
    seedColor: AppColors.teal600,
    brightness: Brightness.light,
  ).copyWith(
    primary: AppColors.teal600,
    onPrimary: AppColors.white,
    secondary: AppColors.coral500,
    onSecondary: AppColors.white,
    surface: AppColors.white,
    surfaceContainerLowest: AppColors.cream,
    surfaceContainerLow: AppColors.paper,
    error: AppColors.error,
    onError: AppColors.white,
  );
  return ThemeData(
    useMaterial3: true,
    colorScheme: scheme,
    scaffoldBackgroundColor: AppColors.cream,
    fontFamily: 'Manrope',
    textTheme: TextTheme(
      displayLarge: AppTypography.display,
      headlineLarge: AppTypography.h1,
      headlineSmall: AppTypography.h2,
      bodyLarge: AppTypography.bodyL,
      bodyMedium: AppTypography.body,
      labelSmall: AppTypography.caption,
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: AppColors.teal600,
        foregroundColor: AppColors.white,
        padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 16),
        shape: RoundedRectangleBorder(borderRadius: AppRadius.r14),
        textStyle: AppTypography.buttonL,
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: AppColors.teal600,
        side: const BorderSide(color: AppColors.teal600, width: 1.5),
        shape: RoundedRectangleBorder(borderRadius: AppRadius.r14),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: AppColors.slate100,
      border: OutlineInputBorder(borderRadius: AppRadius.r14, borderSide: BorderSide.none),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
    ),
    chipTheme: ChipThemeData(
      backgroundColor: AppColors.slate100,
      labelStyle: const TextStyle(fontWeight: FontWeight.w600, fontSize: 12),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      shape: const StadiumBorder(),
    ),
  );
}
```

## 6. Компоненты (01.06)

### `AppButton` — обёртка над всеми вариантами

```dart
enum AppButtonStyle { primary, coral, outline, ghost }

class AppButton extends StatelessWidget {
  final String label;
  final IconData? icon;
  final VoidCallback? onPressed;
  final AppButtonStyle style;
  final bool fullWidth;
  final bool large;
  // ...
  @override Widget build(BuildContext context) {
    // ElevatedButton/OutlinedButton/TextButton по style
  }
}
```

### `AppTextField`, `AppSearchField` — наследуют InputDecoration

### `AppChip` — варианты `tl/cr/sd/outline` через ChipStyle

### `AppCard` — белая карточка с slate200 рамкой и r16

### `RowItem` — paddinged row с avatar + meta + trailing (см. design system 01.06)

### `BottomNav` (5 пунктов + FAB)

```dart
class MainBottomNav extends StatelessWidget {
  final int currentIndex;
  final VoidCallback? onFab;
  // ...
}
```

FAB центральная: квадратный градиент coral500→coral600, тень `AppShadows.fab`, вверх на -22px.

### `MapPin` — кастомная отрисовка пина

```dart
class MapPin extends StatelessWidget {
  final Color color; // teal600 / coral500 / sand700
  final IconData? icon;
  // ...
}
```

### `Skeleton` — shimmer-плейсхолдеры (ST-01)

### `EmptyView`, `ErrorView`, `OfflineBanner`, `Toast`

## 7. Иконография (01.05)

- Outline icons stroke 1.7. Используем **lucide_icons** (есть пакет `lucide_icons: ^0.279`) либо `material_icons` Material 3 (1.5px). Рекомендация: `lucide_icons` — соответствует визуалу дизайна.

## 8. Проверочный список соответствия дизайну

- [ ] Цветовые константы 1-в-1 hex из дизайна
- [ ] Шрифты подключены и применены к h1/body
- [ ] Каждый компонент в `core/design/components/` имеет свой widget-test со скриншотом
- [ ] Все экраны используют только `AppColors`, `AppTypography`, не хардкодят цвета

## 9. Связанные файлы

- Архитектура → `18_FLUTTER_ARCHITECTURE.md`
- Feature-модули → `20_FLUTTER_FEATURES.md`
- Каталог экранов → `02_SCREENS.md`
