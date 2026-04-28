import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';
import '../../../../core/design/tokens.dart';
import '../../domain/entities/trip.dart';
import '../../domain/entities/stop.dart';
import '../providers/stops_provider.dart';
import '../providers/trips_provider.dart';

class TripDetailPage extends ConsumerWidget {
  final Trip trip;
  const TripDetailPage({super.key, required this.trip});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final stopsAsync = ref.watch(stopsProvider(trip.id));
    final routeAsync = ref.watch(tripRouteProvider(trip.id));

    return Scaffold(
      appBar: AppBar(
        title: Text(trip.title),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: stopsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (stops) => Column(
          children: [
            if (stops.isNotEmpty)
              SizedBox(
                height: 220,
                child: _StopsMap(
                  stops: stops,
                  osrmRoute: routeAsync.valueOrNull,
                ),
              ),
            Expanded(
              child: stops.isEmpty
                  ? _EmptyStops(trip: trip)
                  : ListView.separated(
                      padding: const EdgeInsets.all(16),
                      itemCount: stops.length,
                      separatorBuilder: (_, __) => const SizedBox(height: 8),
                      itemBuilder: (ctx, i) => _StopCard(
                        stop: stops[i],
                        tripId: trip.id,
                        index: i + 1,
                      ),
                    ),
            ),
          ],
        ),
      ),
      floatingActionButton: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          FloatingActionButton.small(
            heroTag: 'optimize',
            backgroundColor: AppColors.surface,
            foregroundColor: AppColors.primary,
            tooltip: 'Оптимизировать маршрут',
            child: const Icon(Icons.auto_fix_high),
            onPressed: () => _optimizeRoute(context, ref),
          ),
          const SizedBox(height: 8),
          FloatingActionButton(
            heroTag: 'addStop',
            backgroundColor: AppColors.primary,
            foregroundColor: Colors.white,
            child: const Icon(Icons.add_location),
            onPressed: () => _showAddStopDialog(context, ref),
          ),
        ],
      ),
    );
  }

  Future<void> _optimizeRoute(BuildContext context, WidgetRef ref) async {
    try {
      await ref.read(tripsApiProvider).optimizeRoute(trip.id);
      ref.invalidate(stopsProvider(trip.id));
      ref.invalidate(tripRouteProvider(trip.id));
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Маршрут оптимизирован')),
        );
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка оптимизации: $e')),
        );
      }
    }
  }

  void _showAddStopDialog(BuildContext context, WidgetRef ref) {
    final titleCtrl = TextEditingController();
    final latCtrl = TextEditingController();
    final lngCtrl = TextEditingController();
    final noteCtrl = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Добавить остановку'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(controller: titleCtrl, decoration: const InputDecoration(labelText: 'Название *')),
              const SizedBox(height: 8),
              Row(children: [
                Expanded(child: TextField(controller: latCtrl, decoration: const InputDecoration(labelText: 'Широта *'), keyboardType: TextInputType.number)),
                const SizedBox(width: 8),
                Expanded(child: TextField(controller: lngCtrl, decoration: const InputDecoration(labelText: 'Долгота *'), keyboardType: TextInputType.number)),
              ]),
              const SizedBox(height: 8),
              TextField(controller: noteCtrl, decoration: const InputDecoration(labelText: 'Заметка')),
            ],
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Отмена')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.primary, foregroundColor: Colors.white),
            onPressed: () {
              final title = titleCtrl.text.trim();
              final lat = double.tryParse(latCtrl.text.trim());
              final lng = double.tryParse(lngCtrl.text.trim());
              if (title.isEmpty || lat == null || lng == null) return;
              Navigator.pop(ctx);
              ref.read(stopsProvider(trip.id).notifier).addStop(
                title: title,
                latitude: lat,
                longitude: lng,
                note: noteCtrl.text.trim().isEmpty ? null : noteCtrl.text.trim(),
              );
            },
            child: const Text('Добавить'),
          ),
        ],
      ),
    );
  }
}

class _StopsMap extends StatelessWidget {
  final List<TripStop> stops;
  final List<Map<String, double>>? osrmRoute;
  const _StopsMap({required this.stops, this.osrmRoute});

  @override
  Widget build(BuildContext context) {
    final center = stops.isNotEmpty
        ? LatLng(
            stops.map((s) => s.latitude).reduce((a, b) => a + b) / stops.length,
            stops.map((s) => s.longitude).reduce((a, b) => a + b) / stops.length,
          )
        : const LatLng(55.7558, 37.6173);

    return FlutterMap(
      options: MapOptions(initialCenter: center, initialZoom: 10),
      children: [
        TileLayer(
          urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
          userAgentPackageName: 'com.landmark.app',
        ),
        PolylineLayer(
          polylines: [
            Polyline(
              points: osrmRoute != null && osrmRoute!.isNotEmpty
                  ? osrmRoute!.map((c) => LatLng(c['lat']!, c['lng']!)).toList()
                  : stops.map((s) => LatLng(s.latitude, s.longitude)).toList(),
              color: AppColors.primary,
              strokeWidth: 3,
            ),
          ],
        ),
        MarkerLayer(
          markers: stops.asMap().entries.map((e) => Marker(
            point: LatLng(e.value.latitude, e.value.longitude),
            width: 32, height: 32,
            child: Container(
              decoration: BoxDecoration(
                color: e.value.visitedAt != null ? Colors.green : AppColors.primary,
                shape: BoxShape.circle,
                border: Border.all(color: Colors.white, width: 2),
              ),
              child: Center(
                child: Text('${e.key + 1}', style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold)),
              ),
            ),
          )).toList(),
        ),
      ],
    );
  }
}

class _StopCard extends ConsumerWidget {
  final TripStop stop;
  final String tripId;
  final int index;
  const _StopCard({required this.stop, required this.tripId, required this.index});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final visited = stop.visitedAt != null;
    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: visited ? Colors.green : AppColors.primary,
          foregroundColor: Colors.white,
          child: visited ? const Icon(Icons.check, size: 18) : Text('$index'),
        ),
        title: Text(stop.title, style: const TextStyle(fontWeight: FontWeight.w600)),
        subtitle: stop.note != null ? Text(stop.note!, style: const TextStyle(fontSize: 12)) : null,
        trailing: !visited
            ? TextButton(
                child: const Text('Посещено', style: TextStyle(color: AppColors.primary, fontSize: 12)),
                onPressed: () => ref.read(stopsProvider(tripId).notifier).markVisited(stop.id),
              )
            : const Icon(Icons.check_circle, color: Colors.green),
      ),
    );
  }
}

class _EmptyStops extends StatelessWidget {
  final Trip trip;
  const _EmptyStops({required this.trip});

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.route, size: 64, color: AppColors.textSecondary),
          SizedBox(height: 12),
          Text('Нет остановок', style: TextStyle(fontSize: 18, color: AppColors.textSecondary)),
          SizedBox(height: 8),
          Text('Добавьте точки маршрута', style: TextStyle(color: AppColors.textSecondary)),
        ],
      ),
    );
  }
}
