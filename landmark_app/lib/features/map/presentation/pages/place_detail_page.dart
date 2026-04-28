import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';
import 'package:share_plus/share_plus.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../favorites/presentation/widgets/favorite_button.dart';
import '../../../trips/presentation/providers/trips_provider.dart';
import '../../../trips/domain/entities/trip.dart';
import '../../../trips/presentation/providers/stops_provider.dart';
import '../../domain/entities/place.dart';

class PlaceDetailPage extends ConsumerWidget {
  final Place place;
  const PlaceDetailPage({super.key, required this.place});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 220,
            pinned: true,
            backgroundColor: AppColors.surface,
            foregroundColor: AppColors.textPrimary,
            flexibleSpace: FlexibleSpaceBar(
              background: _PlaceMap(place: place),
            ),
            actions: [
              FavoriteButton(targetType: 'place', targetId: place.id, size: 24),
              IconButton(
                icon: const Icon(Icons.share_outlined),
                onPressed: () => Share.share(
                  '${place.title}${place.city != null ? ', ${place.city}' : ''}\n'
                  'Координаты: ${place.latitude.toStringAsFixed(5)}, ${place.longitude.toStringAsFixed(5)}',
                ),
              ),
            ],
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Заголовок
                  Text(place.title, style: AppTypography.h1),
                  const SizedBox(height: AppSpacing.xs),

                  // Местоположение
                  if (place.city != null || place.country != null)
                    Row(children: [
                      const Icon(Icons.location_on_outlined,
                          size: 16, color: AppColors.textSecondary),
                      const SizedBox(width: 4),
                      Text(
                        [place.city, place.country]
                            .where((s) => s != null)
                            .join(', '),
                        style: AppTypography.bodySmall,
                      ),
                    ]),
                  const SizedBox(height: AppSpacing.md),

                  // Описание
                  if (place.description != null && place.description!.isNotEmpty) ...[
                    Text(place.description!, style: AppTypography.body),
                    const SizedBox(height: AppSpacing.lg),
                  ],

                  // Координаты
                  _InfoRow(
                    icon: Icons.explore_outlined,
                    label:
                        '${place.latitude.toStringAsFixed(5)}, ${place.longitude.toStringAsFixed(5)}',
                  ),
                  const SizedBox(height: AppSpacing.xl),

                  // Кнопка «Добавить в маршрут»
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      icon: const Icon(Icons.add_location_alt_outlined),
                      label: const Text('Добавить в маршрут'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12)),
                      ),
                      onPressed: () => _addToTrip(context, ref),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _addToTrip(BuildContext context, WidgetRef ref) async {
    final trips = await ref.read(tripsApiProvider).listTrips();
    if (!context.mounted) return;

    if (trips.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Сначала создайте поездку')),
      );
      return;
    }

    final selected = await showModalBottomSheet<String>(
      context: context,
      shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (_) => _TripPickerSheet(trips: trips),
    );

    if (selected == null || !context.mounted) return;

    try {
      await ref.read(stopsApiProvider).addStop(
        selected,
        title: place.title,
        latitude: place.latitude,
        longitude: place.longitude,
        placeId: place.id,
      );
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Место добавлено в маршрут')),
        );
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка: $e')),
        );
      }
    }
  }
}

class _PlaceMap extends StatelessWidget {
  final Place place;
  const _PlaceMap({required this.place});

  @override
  Widget build(BuildContext context) {
    return FlutterMap(
      options: MapOptions(
        initialCenter: LatLng(place.latitude, place.longitude),
        initialZoom: 14,
        interactionOptions: const InteractionOptions(flags: InteractiveFlag.none),
      ),
      children: [
        TileLayer(
          urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
          userAgentPackageName: 'com.landmark.app',
        ),
        MarkerLayer(markers: [
          Marker(
            point: LatLng(place.latitude, place.longitude),
            width: 40, height: 40,
            child: Container(
              decoration: BoxDecoration(
                color: AppColors.primary,
                shape: BoxShape.circle,
                border: Border.all(color: Colors.white, width: 2),
                boxShadow: const [
                  BoxShadow(color: Colors.black26, blurRadius: 4, offset: Offset(0, 2))
                ],
              ),
              child: const Icon(Icons.place, color: Colors.white, size: 20),
            ),
          ),
        ]),
      ],
    );
  }
}

class _InfoRow extends StatelessWidget {
  final IconData icon;
  final String label;
  const _InfoRow({required this.icon, required this.label});

  @override
  Widget build(BuildContext context) => Row(
        children: [
          Icon(icon, size: 16, color: AppColors.textSecondary),
          const SizedBox(width: 6),
          Text(label, style: AppTypography.bodySmall),
        ],
      );
}

class _TripPickerSheet extends StatelessWidget {
  final List<Trip> trips;
  const _TripPickerSheet({required this.trips});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const SizedBox(height: 12),
        Container(
          width: 40, height: 4,
          decoration: BoxDecoration(
            color: AppColors.border,
            borderRadius: BorderRadius.circular(2),
          ),
        ),
        const Padding(
          padding: EdgeInsets.all(16),
          child: Text('Выберите поездку',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
        ),
        ...trips.map((t) => ListTile(
              leading: const Icon(Icons.luggage_outlined, color: AppColors.primary),
              title: Text(t.title),
              onTap: () => Navigator.pop(context, t.id),
            )),
        const SizedBox(height: 16),
      ],
    );
  }
}
