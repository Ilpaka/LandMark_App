import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/stops_api.dart';
import '../../domain/entities/stop.dart';
import '../../../../core/networking/api_client.dart';

final stopsApiProvider = Provider((ref) {
  return StopsApi(ref.read(apiClientProvider).dio);
});

final stopsProvider =
    AsyncNotifierProvider.family<StopsNotifier, List<TripStop>, String>(
  StopsNotifier.new,
);

class StopsNotifier extends FamilyAsyncNotifier<List<TripStop>, String> {
  @override
  Future<List<TripStop>> build(String tripId) =>
      ref.read(stopsApiProvider).listStops(tripId);

  Future<void> addStop({
    required String title,
    required double latitude,
    required double longitude,
    String? note,
    String? placeId,
  }) async {
    final stop = await ref.read(stopsApiProvider).addStop(
          arg,
          title: title,
          latitude: latitude,
          longitude: longitude,
          note: note,
          placeId: placeId,
        );
    state = AsyncValue.data([...?state.valueOrNull, stop]);
  }

  Future<void> markVisited(String stopId) async {
    await ref.read(stopsApiProvider).markVisited(arg, stopId);
    ref.invalidateSelf();
  }
}
