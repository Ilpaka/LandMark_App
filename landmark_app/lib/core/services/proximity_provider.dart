import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../networking/api_client.dart';
import '../../features/favorites/data/api/favorites_api.dart';
import '../../features/map/data/api/places_api.dart';
import 'proximity_service.dart';

final proximityServiceProvider = Provider<ProximityService>((ref) {
  return ProximityService();
});

final proximityActiveProvider =
    StateNotifierProvider<ProximityNotifier, bool>((ref) {
  return ProximityNotifier(ref);
});

class ProximityNotifier extends StateNotifier<bool> {
  final Ref _ref;
  ProximityNotifier(this._ref) : super(false);

  Future<void> enable() async {
    final svc = _ref.read(proximityServiceProvider);
    await svc.init();

    // Fetch favorite places and build WatchedPlace list
    final dio = _ref.read(apiClientProvider).dio;
    final favApi = FavoritesApi(dio);
    final placesApi = PlacesApi(dio);

    try {
      final favorites = await favApi.listFavorites(targetType: 'place');
      final watched = <WatchedPlace>[];
      for (final fav in favorites) {
        final id = fav['target_id'] as String?;
        if (id == null) continue;
        try {
          final place = await placesApi.getPlace(id);
          watched.add(WatchedPlace(
            id: place.id,
            title: place.title,
            city: place.city,
            latitude: place.latitude,
            longitude: place.longitude,
          ));
        } catch (_) {}
      }
      final started = await svc.start(watched);
      state = started;
    } catch (_) {
      state = false;
    }
  }

  Future<void> disable() async {
    await _ref.read(proximityServiceProvider).stop();
    state = false;
  }
}
