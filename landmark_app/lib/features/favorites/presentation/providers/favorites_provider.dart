import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/favorites_api.dart';
import '../../../map/domain/entities/place.dart';
import '../../../map/presentation/providers/map_provider.dart';
import '../../../trips/domain/entities/trip.dart';
import '../../../trips/presentation/providers/trips_provider.dart';
import '../../../journal/domain/entities/entry.dart';
import '../../../journal/presentation/providers/journal_provider.dart';
import '../../../../core/networking/api_client.dart';

final favoritesApiProvider = Provider((ref) {
  return FavoritesApi(ref.read(apiClientProvider).dio);
});

// Tracks is-favorite state for a single (targetType, targetId) pair
final isFavoriteProvider =
    StateNotifierProvider.family<FavoriteNotifier, AsyncValue<bool>, (String, String)>(
  (ref, key) => FavoriteNotifier(ref.read(favoritesApiProvider), key.$1, key.$2),
);

class FavoriteNotifier extends StateNotifier<AsyncValue<bool>> {
  final FavoritesApi _api;
  final String _type;
  final String _id;

  FavoriteNotifier(this._api, this._type, this._id) : super(const AsyncValue.loading()) {
    _load();
  }

  Future<void> _load() async {
    try {
      final v = await _api.isFavorite(_type, _id);
      if (mounted) state = AsyncValue.data(v);
    } catch (e, st) {
      if (mounted) state = AsyncValue.error(e, st);
    }
  }

  Future<void> toggle() async {
    final current = state.valueOrNull ?? false;
    state = AsyncValue.data(!current);
    try {
      if (!current) {
        await _api.addFavorite(_type, _id);
      } else {
        await _api.removeFavorite(_type, _id);
      }
    } catch (_) {
      if (mounted) state = AsyncValue.data(current);
    }
  }
}

// ---------------------------------------------------------------------------
// List providers — fetch favorites then resolve full objects in parallel
// ---------------------------------------------------------------------------

final favoritePlacesProvider = FutureProvider.autoDispose<List<Place>>((ref) async {
  final api = ref.read(favoritesApiProvider);
  final repo = ref.read(placesRepositoryProvider);
  final items = await api.listFavorites(targetType: 'place');
  final results = await Future.wait(
    items.map((item) => repo.getPlace(item['target_id'] as String)),
  );
  return results;
});

final favoriteTripsProvider = FutureProvider.autoDispose<List<Trip>>((ref) async {
  final api = ref.read(favoritesApiProvider);
  final tripsApi = ref.read(tripsApiProvider);
  final items = await api.listFavorites(targetType: 'trip');
  final results = await Future.wait(
    items.map((item) => tripsApi.getTrip(item['target_id'] as String)),
  );
  return results;
});

final favoriteEntriesProvider = FutureProvider.autoDispose<List<JournalEntry>>((ref) async {
  final api = ref.read(favoritesApiProvider);
  final journalApi = ref.read(journalApiProvider);
  final items = await api.listFavorites(targetType: 'entry');
  final results = await Future.wait(
    items.map((item) => journalApi.getEntry(item['target_id'] as String)),
  );
  return results;
});
