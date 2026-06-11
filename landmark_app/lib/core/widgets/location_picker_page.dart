import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';

import '../design/tokens.dart';
import 'map_zoom_controls.dart';

/// Полноэкранный выбор координаты на карте.
///
/// API:
///   final point = await LocationPickerPage.show(context, initial: ...);
///
/// Возвращает [LatLng] или `null`, если пользователь закрыл экран без
/// подтверждения.
///
/// UX-паттерн: фиксированный «прицел» в центре экрана + свободно перемещаемая
/// карта под ним. Текущие координаты центра — это координаты, которые вернутся.
/// Тап по карте центрирует её на этой точке (анимированно).
class LocationPickerPage extends StatefulWidget {
  const LocationPickerPage({
    super.key,
    this.initial,
    this.title = 'Выбор точки на карте',
    this.confirmLabel = 'Готово',
  });

  final LatLng? initial;
  final String title;
  final String confirmLabel;

  static Future<LatLng?> show(
    BuildContext context, {
    LatLng? initial,
    String title = 'Выбор точки на карте',
    String confirmLabel = 'Готово',
  }) {
    return Navigator.of(context).push<LatLng>(
      MaterialPageRoute(
        builder: (_) => LocationPickerPage(
          initial: initial,
          title: title,
          confirmLabel: confirmLabel,
        ),
      ),
    );
  }

  @override
  State<LocationPickerPage> createState() => _LocationPickerPageState();
}

class _LocationPickerPageState extends State<LocationPickerPage> {
  static const LatLng _fallbackCenter = LatLng(55.7558, 37.6173); // Москва
  static const double _initialZoom = 12.0;
  static const double _selectedZoom = 14.0;

  final MapController _mapController = MapController();
  StreamSubscription<MapEvent>? _eventSub;

  /// Live-значения. Используем [ValueNotifier] вместо [setState], чтобы
  /// перетаскивание карты не пересобирало [FlutterMap].
  late final ValueNotifier<LatLng> _centerNotifier;
  late final ValueNotifier<double> _zoomNotifier;

  final bool _hasError = false;
  final String? _errorMessage = null;

  @override
  void initState() {
    super.initState();
    final initialCenter = _normalize(widget.initial) ?? _fallbackCenter;
    final initialZoom = widget.initial != null ? _selectedZoom : _initialZoom;
    _centerNotifier = ValueNotifier<LatLng>(initialCenter);
    _zoomNotifier = ValueNotifier<double>(initialZoom);

    // Подписываемся на поток событий карты только после первого кадра, чтобы
    // гарантированно избежать любых build-time реентрей.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      try {
        _eventSub = _mapController.mapEventStream.listen(
          _onMapEvent,
          onError: (Object e, StackTrace st) {
            if (kDebugMode) {
              debugPrint('LocationPicker mapEventStream error: $e\n$st');
            }
          },
        );
      } catch (e, st) {
        if (kDebugMode) {
          debugPrint('LocationPicker subscribe error: $e\n$st');
        }
      }
    });
  }

  @override
  void dispose() {
    _eventSub?.cancel();
    _centerNotifier.dispose();
    _zoomNotifier.dispose();
    _mapController.dispose();
    super.dispose();
  }

  /// Защита от NaN/Infinity и off-globe значений.
  LatLng? _normalize(LatLng? p) {
    if (p == null) return null;
    final lat = p.latitude;
    final lng = p.longitude;
    if (lat.isNaN || lat.isInfinite || lng.isNaN || lng.isInfinite) return null;
    if (lat < -90 || lat > 90 || lng < -180 || lng > 180) return null;
    return p;
  }

  void _onMapEvent(MapEvent event) {
    if (!mounted) return;
    final c = event.camera.center;
    final z = event.camera.zoom;
    if (_centerNotifier.value != c) _centerNotifier.value = c;
    if (_zoomNotifier.value != z) _zoomNotifier.value = z;
  }

  void _onMapTap(TapPosition _, LatLng point) {
    final p = _normalize(point);
    if (p == null) return;
    try {
      _mapController.move(p, _zoomNotifier.value);
    } catch (e) {
      if (kDebugMode) debugPrint('LocationPicker move error: $e');
    }
  }

  void _zoomBy(double delta) {
    final next = (_zoomNotifier.value + delta).clamp(3.0, 18.0);
    try {
      _mapController.move(_centerNotifier.value, next);
    } catch (e) {
      if (kDebugMode) debugPrint('LocationPicker zoom error: $e');
    }
  }

  void _confirm() {
    final c = _normalize(_centerNotifier.value);
    if (c == null) return;
    Navigator.of(context).pop(c);
  }

  @override
  Widget build(BuildContext context) {
    final initialCenter = _normalize(widget.initial) ?? _fallbackCenter;
    final initialZoom = widget.initial != null ? _selectedZoom : _initialZoom;

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: SafeArea(
        top: false,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const _Hint(),
            Expanded(
              child: LayoutBuilder(
                builder: (context, constraints) {
                  // Если по какой-то причине констрейнты схлопнулись —
                  // показываем явный фолбэк, а не белый экран.
                  if (constraints.maxHeight <= 0 || constraints.maxWidth <= 0) {
                    return _ErrorState(
                      message:
                          'Карта не может отобразиться: недостаточно места '
                          'на экране (${constraints.maxWidth.toStringAsFixed(0)}×'
                          '${constraints.maxHeight.toStringAsFixed(0)}).',
                    );
                  }
                  if (_hasError) {
                    return _ErrorState(
                      message:
                          _errorMessage ?? 'Не удалось инициализировать карту',
                    );
                  }
                  return _MapStack(
                    mapController: _mapController,
                    initialCenter: initialCenter,
                    initialZoom: initialZoom,
                    onTap: _onMapTap,
                    onZoomIn: () => _zoomBy(1),
                    onZoomOut: () => _zoomBy(-1),
                  );
                },
              ),
            ),
            _BottomBar(
              centerNotifier: _centerNotifier,
              confirmLabel: widget.confirmLabel,
              onConfirm: _confirm,
            ),
          ],
        ),
      ),
    );
  }
}

class _Hint extends StatelessWidget {
  const _Hint();
  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      color: AppColors.background,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      child: const Row(
        children: [
          Icon(Icons.touch_app, size: 16, color: AppColors.primary),
          SizedBox(width: 6),
          Expanded(
            child: Text(
              'Двигайте карту, чтобы навести «прицел» на нужную точку. '
              'Тап по карте центрирует её на этой точке.',
              style: TextStyle(fontSize: 12, color: AppColors.textPrimary),
            ),
          ),
        ],
      ),
    );
  }
}

class _MapStack extends StatelessWidget {
  const _MapStack({
    required this.mapController,
    required this.initialCenter,
    required this.initialZoom,
    required this.onTap,
    required this.onZoomIn,
    required this.onZoomOut,
  });

  final MapController mapController;
  final LatLng initialCenter;
  final double initialZoom;
  final TapCallback onTap;
  final VoidCallback onZoomIn;
  final VoidCallback onZoomOut;

  @override
  Widget build(BuildContext context) {
    return Stack(
      fit: StackFit.expand,
      children: [
        // Серая «подложка» под тайлами — чтобы при медленной загрузке тайлов
        // пользователь видел контент, а не белый экран.
        Container(color: const Color(0xFFE5E7EB)),
        FlutterMap(
          mapController: mapController,
          options: MapOptions(
            initialCenter: initialCenter,
            initialZoom: initialZoom,
            minZoom: 3,
            maxZoom: 18,
            backgroundColor: const Color(0xFFE5E7EB),
            onTap: onTap,
            interactionOptions: const InteractionOptions(
              flags: InteractiveFlag.all & ~InteractiveFlag.rotate,
            ),
          ),
          children: [
            TileLayer(
              urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
              userAgentPackageName: 'com.landmark.app',
              maxZoom: 19,
              tileProvider: NetworkTileProvider(),
            ),
          ],
        ),
        const IgnorePointer(child: _CenterPin()),
        Positioned(
          right: 12,
          bottom: 12,
          child: MapZoomControls(onZoomIn: onZoomIn, onZoomOut: onZoomOut),
        ),
      ],
    );
  }
}

class _BottomBar extends StatelessWidget {
  const _BottomBar({
    required this.centerNotifier,
    required this.confirmLabel,
    required this.onConfirm,
  });

  final ValueNotifier<LatLng> centerNotifier;
  final String confirmLabel;
  final VoidCallback onConfirm;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: AppColors.surface,
      padding: const EdgeInsets.fromLTRB(16, 8, 12, 12),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'Выбранная точка',
                  style: TextStyle(
                    color: AppColors.textPrimary.withValues(alpha: 0.7),
                    fontSize: 12,
                  ),
                ),
                const SizedBox(height: 2),
                ValueListenableBuilder<LatLng>(
                  valueListenable: centerNotifier,
                  builder: (_, c, __) => Text(
                    '${c.latitude.toStringAsFixed(6)}, '
                    '${c.longitude.toStringAsFixed(6)}',
                    style: const TextStyle(
                      color: AppColors.textPrimary,
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      fontFeatures: [FontFeature.tabularFigures()],
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          ElevatedButton(
            onPressed: onConfirm,
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primary,
              foregroundColor: Colors.white,
              // Явно перекрываем тему: в Row нельзя ставить minWidth=∞,
              // иначе layout падает с BoxConstraints forces an infinite width
              // и весь body Scaffold уходит в белый экран.
              minimumSize: const Size(0, 48),
              padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(10),
              ),
            ),
            child: Text(confirmLabel),
          ),
        ],
      ),
    );
  }
}

class _ErrorState extends StatelessWidget {
  const _ErrorState({required this.message});
  final String message;
  @override
  Widget build(BuildContext context) {
    return Container(
      color: AppColors.background,
      padding: const EdgeInsets.all(24),
      alignment: Alignment.center,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.error_outline, size: 48, color: AppColors.primary),
          const SizedBox(height: 12),
          Text(
            message,
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 14),
          ),
        ],
      ),
    );
  }
}

class _CenterPin extends StatelessWidget {
  const _CenterPin();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Transform.translate(
            offset: const Offset(0, -22),
            child: const Icon(
              Icons.location_on,
              color: AppColors.primary,
              size: 44,
              shadows: [
                Shadow(
                  color: Colors.black38,
                  blurRadius: 4,
                  offset: Offset(0, 2),
                ),
              ],
            ),
          ),
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              color: AppColors.primary,
              shape: BoxShape.circle,
              border: Border.all(color: Colors.white, width: 1.5),
            ),
          ),
        ],
      ),
    );
  }
}

