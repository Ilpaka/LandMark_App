import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/trip.dart';
import '../../data/api/trips_api.dart';
import '../../../../core/networking/api_client.dart';

final tripsApiProvider = Provider((ref) => TripsApi(ref.read(apiClientProvider).dio));

final tripsListProvider =
    AsyncNotifierProvider.autoDispose<TripsListNotifier, List<Trip>>(
  TripsListNotifier.new,
);

class TripsListNotifier extends AutoDisposeAsyncNotifier<List<Trip>> {
  String? _nextCursor;
  bool _hasMore = true;

  @override
  Future<List<Trip>> build() async {
    _nextCursor = null;
    _hasMore = true;
    final result = await ref.read(tripsApiProvider).listTrips();
    _nextCursor = result.nextCursor;
    _hasMore = result.nextCursor != null;
    return result.trips;
  }

  bool get hasMore => _hasMore;

  Future<void> loadMore() async {
    if (!_hasMore || _nextCursor == null) return;
    final current = state.valueOrNull ?? [];
    final result = await ref.read(tripsApiProvider).listTrips(cursor: _nextCursor);
    _nextCursor = result.nextCursor;
    _hasMore = result.nextCursor != null;
    state = AsyncData([...current, ...result.trips]);
  }
}

final publicTripsProvider = FutureProvider.autoDispose<List<Trip>>((ref) async {
  final result = await ref.read(tripsApiProvider).listPublicTrips();
  return result;
});

final tripRouteProvider = FutureProvider.autoDispose.family<List<Map<String, double>>, String>((ref, tripId) {
  return ref.read(tripsApiProvider).getRoute(tripId);
});
