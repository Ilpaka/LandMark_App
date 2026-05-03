import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../map/domain/entities/place.dart';
import '../providers/favorites_provider.dart';

class FavoritesMapPage extends ConsumerStatefulWidget {
  const FavoritesMapPage({super.key});

  @override
  ConsumerState<FavoritesMapPage> createState() => _FavoritesMapPageState();
}

class _FavoritesMapPageState extends ConsumerState<FavoritesMapPage> {
  Place? _selected;

  @override
  Widget build(BuildContext context) {
    final placesAsync = ref.watch(favoritePlacesProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Избранное на карте')),
      body: placesAsync.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (places) {
          if (places.isEmpty) {
            return const Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.map_outlined,
                      size: 64, color: AppColors.textSecondary),
                  SizedBox(height: 16),
                  Text('Нет избранных мест',
                      style: TextStyle(color: AppColors.textSecondary)),
                ],
              ),
            );
          }

          final center = _centerOf(places);

          return Stack(
            children: [
              FlutterMap(
                options: MapOptions(
                  initialCenter: center,
                  initialZoom: places.length == 1 ? 13 : 5,
                  onTap: (_, __) => setState(() => _selected = null),
                ),
                children: [
                  TileLayer(
                    urlTemplate:
                        'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                    userAgentPackageName: 'com.landmark.app',
                  ),
                  MarkerLayer(
                    markers: places
                        .map((p) => Marker(
                              point: LatLng(p.latitude, p.longitude),
                              width: 40,
                              height: 40,
                              child: GestureDetector(
                                onTap: () => setState(() => _selected = p),
                                child: _FavoritePin(
                                    selected: _selected?.id == p.id),
                              ),
                            ))
                        .toList(),
                  ),
                ],
              ),
              if (_selected != null)
                Positioned(
                  bottom: 16,
                  left: 16,
                  right: 16,
                  child: _PlaceCard(
                    place: _selected!,
                    onClose: () => setState(() => _selected = null),
                  ),
                ),
            ],
          );
        },
      ),
    );
  }

  LatLng _centerOf(List<Place> places) {
    if (places.length == 1)
      return LatLng(places.first.latitude, places.first.longitude);
    final lat =
        places.map((p) => p.latitude).reduce((a, b) => a + b) / places.length;
    final lng =
        places.map((p) => p.longitude).reduce((a, b) => a + b) / places.length;
    return LatLng(lat, lng);
  }
}

class _FavoritePin extends StatelessWidget {
  final bool selected;
  const _FavoritePin({this.selected = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: selected ? AppColors.accent : Colors.red,
        shape: BoxShape.circle,
        border: Border.all(color: Colors.white, width: 2),
        boxShadow: const [
          BoxShadow(color: Colors.black26, blurRadius: 4, offset: Offset(0, 2))
        ],
      ),
      child: const Icon(Icons.favorite, color: Colors.white, size: 20),
    );
  }
}

class _PlaceCard extends StatelessWidget {
  final Place place;
  final VoidCallback onClose;

  const _PlaceCard({required this.place, required this.onClose});

  @override
  Widget build(BuildContext context) {
    return Material(
      elevation: 8,
      borderRadius: BorderRadius.circular(16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: AppColors.primary.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(Icons.place, color: AppColors.primary),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(place.title,
                      style: AppTypography.h3,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis),
                  if (place.city != null)
                    Text(place.city!, style: AppTypography.bodySmall),
                ],
              ),
            ),
            IconButton(icon: const Icon(Icons.close), onPressed: onClose),
          ],
        ),
      ),
    );
  }
}
