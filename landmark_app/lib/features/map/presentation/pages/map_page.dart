import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:geolocator/geolocator.dart';
import 'package:latlong2/latlong.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/widgets/map_zoom_controls.dart';
import '../providers/map_provider.dart';
import '../../domain/entities/place.dart';
import '../../../favorites/presentation/widgets/favorite_button.dart';
import 'place_detail_page.dart';
import 'suggest_place_page.dart';
import '../../../search/presentation/pages/search_page.dart';

class MapPage extends ConsumerStatefulWidget {
  const MapPage({super.key});

  @override
  ConsumerState<MapPage> createState() => _MapPageState();
}

class _MapPageState extends ConsumerState<MapPage> {
  final _mapController = MapController();
  Timer? _debounce;
  Place? _selectedPlace;
  bool _locating = false;

  static const _minZoom = 2.0;
  static const _maxZoom = 18.0;

  @override
  void initState() {
    super.initState();
    // Загрузить места при старте
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(mapBBoxProvider.notifier).state = const PlacesQuery();
    });
  }

  @override
  void dispose() {
    _debounce?.cancel();
    super.dispose();
  }

  /// Вызывается на ЛЮБОЕ изменение камеры (жест, кнопки зума, programmatic
  /// move) — в отличие от подписки на отдельные MapEvent'ы, которая
  /// пропускала pinch-zoom и programmatic-перемещения, оставляя bbox
  /// устаревшим.
  void _onPositionChanged(MapPosition pos, bool hasGesture) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 400), () {
      if (!mounted) return;
      final bounds = pos.bounds ?? _mapController.camera.visibleBounds;
      final category = ref.read(selectedCategoryProvider);
      ref.read(mapBBoxProvider.notifier).state = PlacesQuery(
        south: bounds.south,
        west: bounds.west,
        north: bounds.north,
        east: bounds.east,
        categorySlug: category,
      );
    });
  }

  void _zoomBy(double delta) {
    final camera = _mapController.camera;
    _mapController.move(
      camera.center,
      (camera.zoom + delta).clamp(_minZoom, _maxZoom),
    );
  }

  Future<void> _goToMyLocation() async {
    setState(() => _locating = true);
    try {
      final serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        _showMessage('Включите службы геолокации в настройках устройства');
        return;
      }
      var perm = await Geolocator.checkPermission();
      if (perm == LocationPermission.denied) {
        perm = await Geolocator.requestPermission();
      }
      if (perm == LocationPermission.denied ||
          perm == LocationPermission.deniedForever) {
        _showMessage('Нет доступа к геолокации');
        return;
      }
      final pos = await Geolocator.getCurrentPosition();
      if (!mounted) return;
      _mapController.move(LatLng(pos.latitude, pos.longitude), 14);
    } catch (_) {
      _showMessage('Не удалось определить местоположение');
    } finally {
      if (mounted) setState(() => _locating = false);
    }
  }

  void _showMessage(String text) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(text)));
  }

  void _showPlacePreview(Place place) {
    setState(() => _selectedPlace = place);
  }

  Future<void> _addPlace() async {
    final created = await Navigator.push<Place?>(
      context,
      MaterialPageRoute(
        builder: (_) => SuggestPlacePage(
          initialCoords: _mapController.camera.center,
        ),
      ),
    );
    if (!mounted) return;
    // Сбрасываем кэш списка мест: могла появиться новая приватная точка
    // (публикуется сразу) или публичный draft (не виден до approve).
    ref.invalidate(placesProvider);
    if (created != null) {
      // У новой точки ещё нет категорий — с активным фильтром она не
      // попала бы в выдачу, и пользователь решил бы, что точка пропала.
      ref.read(selectedCategoryProvider.notifier).state = null;
      _mapController.move(
        LatLng(created.latitude, created.longitude),
        15,
      );
      _showPlacePreview(created);
    }
  }

  @override
  Widget build(BuildContext context) {
    final query = ref.watch(mapBBoxProvider);
    final placesAsync = query != null
        ? ref.watch(placesProvider(query))
        : const AsyncValue<List<Place>>.data([]);

    return Scaffold(
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _addPlace,
        backgroundColor: AppColors.primary,
        foregroundColor: Colors.white,
        icon: const Icon(Icons.add_location_alt),
        label: const Text('Добавить место'),
      ),
      body: Stack(
        children: [
          FlutterMap(
            mapController: _mapController,
            options: MapOptions(
              initialCenter: const LatLng(55.7558, 37.6173), // Москва
              initialZoom: 5,
              minZoom: _minZoom,
              maxZoom: _maxZoom,
              // Явно включаем все жесты, кроме поворота карты: вращение
              // случайными жестами дезориентирует, а пользы для POI-карты
              // не даёт.
              interactionOptions: const InteractionOptions(
                flags: InteractiveFlag.all & ~InteractiveFlag.rotate,
              ),
              onTap: (_, __) => setState(() => _selectedPlace = null),
              onPositionChanged: _onPositionChanged,
            ),
            children: [
              TileLayer(
                urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                userAgentPackageName: 'com.landmark.app',
                maxZoom: 19,
              ),
              MarkerLayer(
                markers: placesAsync.valueOrNull
                        ?.map((place) => Marker(
                              point: LatLng(place.latitude, place.longitude),
                              width: 40,
                              height: 40,
                              child: GestureDetector(
                                onTap: () => _showPlacePreview(place),
                                child: _PlacePin(
                                  selected: _selectedPlace?.id == place.id,
                                  isPrivate: place.isPrivate,
                                ),
                              ),
                            ))
                        .toList() ??
                    [],
              ),
              const SimpleAttributionWidget(
                source: Text('OpenStreetMap'),
              ),
            ],
          ),

          // Search bar сверху
          Positioned(
            top: MediaQuery.of(context).padding.top + 8,
            left: 16,
            right: 16,
            child: _SearchBarTap(onTap: () async {
              final place = await Navigator.push<Place>(
                context,
                MaterialPageRoute(builder: (_) => const SearchPage()),
              );
              if (place != null && mounted) {
                _mapController.move(
                    LatLng(place.latitude, place.longitude), 14);
                _showPlacePreview(place);
              }
            }),
          ),

          // Фильтр категорий
          Positioned(
            top: MediaQuery.of(context).padding.top + 64,
            left: 0,
            right: 0,
            child: const _CategoryFilter(),
          ),

          // Кнопки зума и геолокации
          Positioned(
            right: 12,
            bottom: _selectedPlace != null ? 130 : 96,
            child: MapZoomControls(
              onZoomIn: () => _zoomBy(1),
              onZoomOut: () => _zoomBy(-1),
              onLocate: _goToMyLocation,
              locating: _locating,
            ),
          ),

          // Индикатор загрузки
          if (placesAsync.isLoading)
            Positioned(
              top: MediaQuery.of(context).padding.top + 112,
              left: 0,
              right: 0,
              child: const Center(child: _LoadingPill()),
            ),

          // Ошибка загрузки мест: без этого падение запроса оставляло
          // пользователя наедине с пустой картой без объяснений.
          if (placesAsync.hasError && !placesAsync.isLoading)
            Positioned(
              top: MediaQuery.of(context).padding.top + 112,
              left: 0,
              right: 0,
              child: Center(
                child: _ErrorPill(
                  onRetry: () => ref.invalidate(placesProvider),
                ),
              ),
            ),

          // Preview нижней карточки
          if (_selectedPlace != null)
            Positioned(
              bottom: 16,
              left: 16,
              right: 16,
              child: _PlacePreviewCard(
                place: _selectedPlace!,
                onClose: () => setState(() => _selectedPlace = null),
              ),
            ),
        ],
      ),
    );
  }
}

class _LoadingPill extends StatelessWidget {
  const _LoadingPill();

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
      elevation: 2,
      borderRadius: BorderRadius.circular(20),
      child: const Padding(
        padding: EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            SizedBox(
              width: 14,
              height: 14,
              child: CircularProgressIndicator(
                strokeWidth: 2,
                color: AppColors.primary,
              ),
            ),
            SizedBox(width: 8),
            Text('Обновляем места…', style: TextStyle(fontSize: 13)),
          ],
        ),
      ),
    );
  }
}

class _ErrorPill extends StatelessWidget {
  const _ErrorPill({required this.onRetry});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
      elevation: 2,
      borderRadius: BorderRadius.circular(20),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(14, 4, 4, 4),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off, size: 16, color: AppColors.error),
            const SizedBox(width: 8),
            const Text('Не удалось загрузить места',
                style: TextStyle(fontSize: 13)),
            TextButton(
              onPressed: onRetry,
              child: const Text('Повторить'),
            ),
          ],
        ),
      ),
    );
  }
}

class _PlacePin extends StatelessWidget {
  final bool selected;
  final bool isPrivate;
  const _PlacePin({this.selected = false, this.isPrivate = false});

  @override
  Widget build(BuildContext context) {
    final color = selected
        ? AppColors.accent
        : (isPrivate ? AppColors.textSecondary : AppColors.primary);
    return AnimatedScale(
      scale: selected ? 1.15 : 1,
      duration: const Duration(milliseconds: 150),
      child: Container(
        decoration: BoxDecoration(
          color: color,
          shape: BoxShape.circle,
          border: Border.all(color: Colors.white, width: 2),
          boxShadow: const [
            BoxShadow(
                color: Colors.black26, blurRadius: 4, offset: Offset(0, 2))
          ],
        ),
        child: Icon(
          isPrivate ? Icons.lock : Icons.place,
          color: Colors.white,
          size: 18,
        ),
      ),
    );
  }
}

class _SearchBarTap extends StatelessWidget {
  final VoidCallback onTap;
  const _SearchBarTap({required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Material(
        elevation: 4,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
          ),
          child: const Row(
            children: [
              Icon(Icons.search, color: AppColors.textSecondary),
              SizedBox(width: 12),
              Text('Поиск мест...',
                  style:
                      TextStyle(color: AppColors.textSecondary, fontSize: 16)),
            ],
          ),
        ),
      ),
    );
  }
}

class _CategoryFilter extends ConsumerWidget {
  const _CategoryFilter();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cats = ref.watch(categoriesProvider);
    final selected = ref.watch(selectedCategoryProvider);

    return cats.when(
      data: (list) => SizedBox(
        height: 40,
        child: ListView.separated(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          itemCount: list.length,
          separatorBuilder: (_, __) => const SizedBox(width: 8),
          itemBuilder: (_, i) {
            final cat = list[i];
            final isSelected = selected == cat.slug;
            return GestureDetector(
              onTap: () {
                ref.read(selectedCategoryProvider.notifier).state =
                    isSelected ? null : cat.slug;
                final bbox = ref.read(mapBBoxProvider);
                ref.read(mapBBoxProvider.notifier).state = PlacesQuery(
                  south: bbox?.south,
                  west: bbox?.west,
                  north: bbox?.north,
                  east: bbox?.east,
                  categorySlug: isSelected ? null : cat.slug,
                );
              },
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 150),
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: isSelected ? AppColors.primary : Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: const [
                    BoxShadow(color: Colors.black12, blurRadius: 4)
                  ],
                ),
                child: Text(
                  cat.title,
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                    color: isSelected ? Colors.white : AppColors.textPrimary,
                  ),
                ),
              ),
            );
          },
        ),
      ),
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
    );
  }
}

class _PlacePreviewCard extends StatelessWidget {
  final Place place;
  final VoidCallback onClose;

  const _PlacePreviewCard({required this.place, required this.onClose});

  @override
  Widget build(BuildContext context) {
    return Material(
      elevation: 8,
      borderRadius: BorderRadius.circular(16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            // Иконка места
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: (place.isPrivate
                        ? AppColors.textSecondary
                        : AppColors.primary)
                    .withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                place.isPrivate ? Icons.lock : Icons.place,
                color: place.isPrivate
                    ? AppColors.textSecondary
                    : AppColors.primary,
                size: 28,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    place.title,
                    style: AppTypography.h3,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  if (place.isPrivate)
                    const Text('Приватная точка',
                        style: TextStyle(
                            fontSize: 12, color: AppColors.textSecondary))
                  else if (place.city != null)
                    Text(place.city!, style: AppTypography.bodySmall),
                ],
              ),
            ),
            FavoriteButton(targetType: 'place', targetId: place.id, size: 26),
            IconButton(
              icon: const Icon(Icons.open_in_new, size: 20),
              tooltip: 'Подробнее',
              onPressed: () => Navigator.push<void>(
                context,
                MaterialPageRoute(
                    builder: (_) => PlaceDetailPage(place: place)),
              ),
            ),
            IconButton(icon: const Icon(Icons.close), onPressed: onClose),
          ],
        ),
      ),
    );
  }
}
