import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/widgets/location_picker_page.dart';
import '../../domain/entities/trip.dart';
import '../../domain/entities/stop.dart';
import '../providers/stops_provider.dart';
import '../providers/trips_provider.dart';
import 'trip_map_page.dart';

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
        actions: [
          IconButton(
            icon: const Icon(Icons.map_outlined),
            tooltip: 'Карта маршрута',
            onPressed: () => Navigator.push<void>(
              context,
              MaterialPageRoute(builder: (_) => TripMapPage(trip: trip)),
            ),
          ),
          PopupMenuButton<String>(
            onSelected: (v) => _onMenuAction(context, ref, v),
            itemBuilder: (_) => [
              if (trip.status == 'planned')
                const PopupMenuItem(
                    value: 'start', child: Text('Начать поездку')),
              if (trip.status == 'in_progress')
                const PopupMenuItem(
                    value: 'complete', child: Text('Завершить поездку')),
              if (trip.status == 'completed' || trip.status == 'in_progress')
                const PopupMenuItem(
                    value: 'plan', child: Text('Вернуть в запланированные')),
              const PopupMenuItem(
                value: 'delete',
                child: Text('Удалить', style: TextStyle(color: Colors.red)),
              ),
            ],
          ),
        ],
      ),
      body: stopsAsync.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
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

  Future<void> _onMenuAction(
      BuildContext context, WidgetRef ref, String action) async {
    if (action == 'delete') {
      final confirm = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('Удалить поездку?'),
          content: const Text('Все остановки будут удалены.'),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(ctx, false),
                child: const Text('Отмена')),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.red, foregroundColor: Colors.white),
              onPressed: () => Navigator.pop(ctx, true),
              child: const Text('Удалить'),
            ),
          ],
        ),
      );
      if (confirm == true && context.mounted) {
        await ref.read(tripsApiProvider).deleteTrip(trip.id);
        ref.invalidate(tripsListProvider);
        if (context.mounted) Navigator.pop(context);
      }
      return;
    }
    final newStatus = switch (action) {
      'start' => 'in_progress',
      'complete' => 'completed',
      'plan' => 'planned',
      _ => null,
    };
    if (newStatus == null) return;
    try {
      await ref.read(tripsApiProvider).updateTrip(trip.id, status: newStatus);
      ref.invalidate(tripsListProvider);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(_statusLabel(newStatus))),
        );
        Navigator.pop(context);
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка: $e')),
        );
      }
    }
  }

  String _statusLabel(String status) => switch (status) {
        'in_progress' => 'Поездка началась!',
        'completed' => 'Поездка завершена',
        'planned' => 'Поездка возвращена в запланированные',
        _ => 'Статус обновлён',
      };

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
    // Подбираем стартовый центр карты по уже добавленным остановкам.
    final stops = ref.read(stopsProvider(trip.id)).maybeWhen(
          data: (data) => data,
          orElse: () => const <TripStop>[],
        );
    final initial = stops.isNotEmpty
        ? LatLng(stops.first.latitude, stops.first.longitude)
        : null;
    showDialog(
      context: context,
      builder: (_) => _AddStopDialog(
        initialCenter: initial,
        onSubmit: ({
          required String title,
          required LatLng coords,
          String? note,
        }) {
          ref.read(stopsProvider(trip.id).notifier).addStop(
                title: title,
                latitude: coords.latitude,
                longitude: coords.longitude,
                note: note,
              );
        },
      ),
    );
  }
}

class _AddStopDialog extends StatefulWidget {
  const _AddStopDialog({required this.onSubmit, this.initialCenter});

  final LatLng? initialCenter;
  final void Function({
    required String title,
    required LatLng coords,
    String? note,
  }) onSubmit;

  @override
  State<_AddStopDialog> createState() => _AddStopDialogState();
}

class _AddStopDialogState extends State<_AddStopDialog> {
  final _titleCtrl = TextEditingController();
  final _noteCtrl = TextEditingController();
  LatLng? _coords;
  bool _showError = false;

  @override
  void dispose() {
    _titleCtrl.dispose();
    _noteCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickCoords() async {
    final picked = await LocationPickerPage.show(
      context,
      initial: _coords ?? widget.initialCenter,
      title: 'Где будет остановка?',
    );
    if (!mounted) return;
    if (picked != null) {
      setState(() {
        _coords = picked;
        _showError = false;
      });
    }
  }

  void _confirm() {
    final title = _titleCtrl.text.trim();
    final coords = _coords;
    if (title.isEmpty || coords == null) {
      setState(() => _showError = title.isEmpty || coords == null);
      return;
    }
    Navigator.pop(context);
    widget.onSubmit(
      title: title,
      coords: coords,
      note: _noteCtrl.text.trim().isEmpty ? null : _noteCtrl.text.trim(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final hasCoords = _coords != null;
    final coordsBorder = _showError && !hasCoords
        ? Colors.red
        : hasCoords
            ? AppColors.primary
            : AppColors.border;

    return AlertDialog(
      title: const Text('Добавить остановку'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
              controller: _titleCtrl,
              decoration: const InputDecoration(labelText: 'Название *'),
            ),
            const SizedBox(height: 12),
            InkWell(
              onTap: _pickCoords,
              borderRadius: BorderRadius.circular(10),
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: coordsBorder, width: 1),
                ),
                child: Row(
                  children: [
                    Icon(
                      hasCoords
                          ? Icons.location_on
                          : Icons.add_location_alt_outlined,
                      size: 20,
                      color: hasCoords
                          ? AppColors.primary
                          : AppColors.textPrimary.withValues(alpha: 0.6),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            hasCoords
                                ? 'Точка на карте *'
                                : 'Выбрать точку на карте *',
                            style: TextStyle(
                              fontSize: 11,
                              color:
                                  AppColors.textPrimary.withValues(alpha: 0.7),
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            hasCoords
                                ? '${_coords!.latitude.toStringAsFixed(5)}, '
                                    '${_coords!.longitude.toStringAsFixed(5)}'
                                : 'Координаты не выбраны',
                            style: TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                              color: hasCoords
                                  ? AppColors.textPrimary
                                  : AppColors.textPrimary
                                      .withValues(alpha: 0.5),
                              fontFeatures: const [
                                FontFeature.tabularFigures()
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    Icon(
                      hasCoords ? Icons.edit : Icons.chevron_right,
                      size: 18,
                      color: AppColors.textPrimary.withValues(alpha: 0.5),
                    ),
                  ],
                ),
              ),
            ),
            if (_showError && !hasCoords) ...[
              const SizedBox(height: 6),
              const Text(
                'Поставьте метку на карте',
                style: TextStyle(fontSize: 12, color: Colors.red),
              ),
            ],
            const SizedBox(height: 12),
            TextField(
              controller: _noteCtrl,
              decoration: const InputDecoration(labelText: 'Заметка'),
              maxLines: 2,
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Отмена'),
        ),
        ElevatedButton(
          style: ElevatedButton.styleFrom(
            backgroundColor: AppColors.primary,
            foregroundColor: Colors.white,
          ),
          onPressed: _confirm,
          child: const Text('Добавить'),
        ),
      ],
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
            stops.map((s) => s.longitude).reduce((a, b) => a + b) /
                stops.length,
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
          markers: stops
              .asMap()
              .entries
              .map((e) => Marker(
                    point: LatLng(e.value.latitude, e.value.longitude),
                    width: 32,
                    height: 32,
                    child: Container(
                      decoration: BoxDecoration(
                        color: e.value.visitedAt != null
                            ? Colors.green
                            : AppColors.primary,
                        shape: BoxShape.circle,
                        border: Border.all(color: Colors.white, width: 2),
                      ),
                      child: Center(
                        child: Text('${e.key + 1}',
                            style: const TextStyle(
                                color: Colors.white,
                                fontSize: 12,
                                fontWeight: FontWeight.bold)),
                      ),
                    ),
                  ))
              .toList(),
        ),
      ],
    );
  }
}

class _StopCard extends ConsumerWidget {
  final TripStop stop;
  final String tripId;
  final int index;
  const _StopCard(
      {required this.stop, required this.tripId, required this.index});

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
        title: Text(stop.title,
            style: const TextStyle(fontWeight: FontWeight.w600)),
        subtitle: stop.note != null
            ? Text(stop.note!, style: const TextStyle(fontSize: 12))
            : null,
        trailing: !visited
            ? TextButton(
                child: const Text('Посещено',
                    style: TextStyle(color: AppColors.primary, fontSize: 12)),
                onPressed: () => ref
                    .read(stopsProvider(tripId).notifier)
                    .markVisited(stop.id),
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
          Text('Нет остановок',
              style: TextStyle(fontSize: 18, color: AppColors.textSecondary)),
          SizedBox(height: 8),
          Text('Добавьте точки маршрута',
              style: TextStyle(color: AppColors.textSecondary)),
        ],
      ),
    );
  }
}
