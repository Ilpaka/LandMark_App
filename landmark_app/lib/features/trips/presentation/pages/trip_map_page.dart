import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';
import '../../../../core/design/tokens.dart';
import '../../domain/entities/stop.dart';
import '../../domain/entities/trip.dart';
import '../providers/stops_provider.dart';
import '../providers/trips_provider.dart';

class TripMapPage extends ConsumerStatefulWidget {
  final Trip trip;
  const TripMapPage({super.key, required this.trip});

  @override
  ConsumerState<TripMapPage> createState() => _TripMapPageState();
}

class _TripMapPageState extends ConsumerState<TripMapPage> {
  final _mapController = MapController();
  int? _selectedIdx;

  void _fitBounds(List<TripStop> stops) {
    if (stops.isEmpty) return;
    if (stops.length == 1) {
      _mapController.move(
        LatLng(stops.first.latitude, stops.first.longitude),
        14,
      );
      return;
    }
    final lats = stops.map((s) => s.latitude);
    final lngs = stops.map((s) => s.longitude);
    final bounds = LatLngBounds(
      LatLng(lats.reduce((a, b) => a < b ? a : b),
          lngs.reduce((a, b) => a < b ? a : b)),
      LatLng(lats.reduce((a, b) => a > b ? a : b),
          lngs.reduce((a, b) => a > b ? a : b)),
    );
    _mapController.fitCamera(
      CameraFit.bounds(bounds: bounds, padding: const EdgeInsets.all(60)),
    );
  }

  @override
  Widget build(BuildContext context) {
    final stopsAsync = ref.watch(stopsProvider(widget.trip.id));
    final routeAsync = ref.watch(tripRouteProvider(widget.trip.id));

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.trip.title),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          stopsAsync.when(
            data: (stops) => IconButton(
              icon: const Icon(Icons.fit_screen),
              tooltip: 'Показать все точки',
              onPressed: () => _fitBounds(stops),
            ),
            loading: () => const SizedBox.shrink(),
            error: (_, __) => const SizedBox.shrink(),
          ),
        ],
      ),
      body: stopsAsync.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (stops) {
          final rawRoute = routeAsync.valueOrNull;
          final route =
              rawRoute?.map((p) => LatLng(p['lat']!, p['lng']!)).toList();
          return Stack(
            children: [
              FlutterMap(
                mapController: _mapController,
                options: MapOptions(
                  initialCenter: stops.isNotEmpty
                      ? LatLng(stops.first.latitude, stops.first.longitude)
                      : const LatLng(55.75, 37.62),
                  initialZoom: 12,
                  onMapReady: () => WidgetsBinding.instance
                      .addPostFrameCallback((_) => _fitBounds(stops)),
                ),
                children: [
                  TileLayer(
                    urlTemplate:
                        'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                    userAgentPackageName: 'com.wanderlog.app',
                  ),
                  // Route polyline
                  if (route != null && route.isNotEmpty)
                    PolylineLayer(polylines: [
                      Polyline(
                        points: route,
                        color: AppColors.primary,
                        strokeWidth: 3.5,
                      ),
                    ]),
                  // Connecting lines (straight fallback)
                  if ((route == null || route.isEmpty) && stops.length > 1)
                    PolylineLayer(polylines: [
                      Polyline(
                        points: stops
                            .map((s) => LatLng(s.latitude, s.longitude))
                            .toList(),
                        color: AppColors.primary.withValues(alpha: 0.5),
                        strokeWidth: 2,
                        isDotted: true,
                      ),
                    ]),
                  // Stop markers
                  MarkerLayer(
                    markers: [
                      for (var i = 0; i < stops.length; i++)
                        Marker(
                          point: LatLng(stops[i].latitude, stops[i].longitude),
                          width: 40,
                          height: 40,
                          child: GestureDetector(
                            onTap: () => setState(() =>
                                _selectedIdx = _selectedIdx == i ? null : i),
                            child: _StopPin(
                              index: i + 1,
                              isSelected: _selectedIdx == i,
                              visitedAt: stops[i].visitedAt,
                            ),
                          ),
                        ),
                    ],
                  ),
                ],
              ),
              // Selected stop info card
              if (_selectedIdx != null && _selectedIdx! < stops.length)
                Positioned(
                  left: 16,
                  right: 16,
                  bottom: 24,
                  child: _StopInfoCard(
                    stop: stops[_selectedIdx!],
                    index: _selectedIdx! + 1,
                    onClose: () => setState(() => _selectedIdx = null),
                  ),
                ),
            ],
          );
        },
      ),
    );
  }
}

class _StopPin extends StatelessWidget {
  final int index;
  final bool isSelected;
  final DateTime? visitedAt;
  const _StopPin(
      {required this.index, required this.isSelected, this.visitedAt});

  @override
  Widget build(BuildContext context) {
    final visited = visitedAt != null;
    final bg = isSelected
        ? AppColors.primary
        : visited
            ? Colors.green
            : AppColors.primary.withValues(alpha: 0.8);

    return Container(
      decoration: BoxDecoration(
        color: bg,
        shape: BoxShape.circle,
        border: Border.all(color: Colors.white, width: 2),
        boxShadow: const [
          BoxShadow(color: Colors.black26, blurRadius: 4, offset: Offset(0, 2))
        ],
      ),
      alignment: Alignment.center,
      child: Text(
        '$index',
        style: const TextStyle(
            color: Colors.white, fontSize: 13, fontWeight: FontWeight.w700),
      ),
    );
  }
}

class _StopInfoCard extends StatelessWidget {
  final TripStop stop;
  final int index;
  final VoidCallback onClose;
  const _StopInfoCard(
      {required this.stop, required this.index, required this.onClose});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 6,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 12, 8, 12),
        child: Row(
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: const BoxDecoration(
                color: AppColors.primary,
                shape: BoxShape.circle,
              ),
              alignment: Alignment.center,
              child: Text('$index',
                  style: const TextStyle(
                      color: Colors.white, fontWeight: FontWeight.w700)),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(stop.title,
                      style: const TextStyle(
                          fontWeight: FontWeight.w600, fontSize: 15),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis),
                  if (stop.note != null && stop.note!.isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(stop.note!,
                        style: const TextStyle(
                            fontSize: 13, color: AppColors.textSecondary),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis),
                  ],
                  if (stop.visitedAt != null)
                    const Padding(
                      padding: EdgeInsets.only(top: 2),
                      child: Row(
                        children: [
                          Icon(Icons.check_circle,
                              size: 14, color: Colors.green),
                          SizedBox(width: 4),
                          Text('Посещено',
                              style:
                                  TextStyle(fontSize: 12, color: Colors.green)),
                        ],
                      ),
                    ),
                ],
              ),
            ),
            IconButton(
              icon: const Icon(Icons.close, size: 20),
              onPressed: onClose,
              color: AppColors.textSecondary,
            ),
          ],
        ),
      ),
    );
  }
}
