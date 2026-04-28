import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../providers/map_provider.dart';
import '../../domain/entities/place.dart';
import '../../../favorites/presentation/widgets/favorite_button.dart';
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

  void _onMapMoved(MapCamera camera) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 500), () {
      if (!mounted) return;
      final bounds = camera.visibleBounds;
      final category = ref.read(selectedCategoryProvider);
      ref.read(mapBBoxProvider.notifier).state = PlacesQuery(
        south: bounds.south, west: bounds.west, north: bounds.north, east: bounds.east,
        categorySlug: category,
      );
    });
  }

  void _showPlacePreview(Place place) {
    setState(() => _selectedPlace = place);
  }

  @override
  Widget build(BuildContext context) {
    final query = ref.watch(mapBBoxProvider);
    final placesAsync = query != null
        ? ref.watch(placesProvider(query))
        : const AsyncValue<List<Place>>.data([]);

    return Scaffold(
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => Navigator.push(
          context,
          MaterialPageRoute(builder: (_) => const SuggestPlacePage()),
        ),
        backgroundColor: AppColors.primary,
        foregroundColor: Colors.white,
        icon: const Icon(Icons.add_location_alt),
        label: const Text('Предложить место'),
      ),
      body: Stack(
        children: [
          FlutterMap(
            mapController: _mapController,
            options: MapOptions(
              initialCenter: const LatLng(55.7558, 37.6173), // Москва
              initialZoom: 5,
              onMapEvent: (event) {
                if (event is MapEventMoveEnd || event is MapEventScrollWheelZoom) {
                  _onMapMoved(_mapController.camera);
                }
              },
            ),
            children: [
              TileLayer(
                urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                userAgentPackageName: 'com.landmark.app',
              ),
              MarkerLayer(
                markers: placesAsync.valueOrNull?.map((place) => Marker(
                  point: LatLng(place.latitude, place.longitude),
                  width: 40,
                  height: 40,
                  child: GestureDetector(
                    onTap: () => _showPlacePreview(place),
                    child: _PlacePin(selected: _selectedPlace?.id == place.id),
                  ),
                )).toList() ?? [],
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
                _mapController.move(LatLng(place.latitude, place.longitude), 14);
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

          // Индикатор загрузки
          if (placesAsync.isLoading)
            const Positioned(
              bottom: 100,
              left: 0,
              right: 0,
              child: Center(child: CircularProgressIndicator(color: AppColors.primary)),
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

class _PlacePin extends StatelessWidget {
  final bool selected;
  const _PlacePin({this.selected = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: selected ? AppColors.accent : AppColors.primary,
        shape: BoxShape.circle,
        border: Border.all(color: Colors.white, width: 2),
        boxShadow: const [BoxShadow(color: Colors.black26, blurRadius: 4, offset: Offset(0, 2))],
      ),
      child: const Icon(Icons.place, color: Colors.white, size: 20),
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
              Text('Поиск мест...', style: TextStyle(color: AppColors.textSecondary, fontSize: 16)),
            ],
          ),
        ),
      ),
    );
  }
}

class _SearchBar extends StatefulWidget {
  final ValueChanged<String> onSearch;
  const _SearchBar({required this.onSearch});

  @override
  State<_SearchBar> createState() => _SearchBarState();
}

class _SearchBarState extends State<_SearchBar> {
  final _ctrl = TextEditingController();

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Material(
      elevation: 4,
      borderRadius: BorderRadius.circular(12),
      child: TextField(
        controller: _ctrl,
        decoration: InputDecoration(
          hintText: 'Поиск мест...',
          prefixIcon: const Icon(Icons.search, color: AppColors.textSecondary),
          suffixIcon: _ctrl.text.isNotEmpty
              ? IconButton(
                  icon: const Icon(Icons.clear),
                  onPressed: () {
                    _ctrl.clear();
                    widget.onSearch('');
                    setState(() {});
                  },
                )
              : null,
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        ),
        onSubmitted: widget.onSearch,
        onChanged: (_) => setState(() {}),
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
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: isSelected ? AppColors.primary : Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4)],
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
                color: AppColors.primary.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(Icons.place, color: AppColors.primary, size: 28),
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
                  if (place.city != null)
                    Text(place.city!, style: AppTypography.bodySmall),
                ],
              ),
            ),
            FavoriteButton(targetType: 'place', targetId: place.id, size: 26),
            const SizedBox(width: 4),
            IconButton(icon: const Icon(Icons.close), onPressed: onClose),
          ],
        ),
      ),
    );
  }
}
