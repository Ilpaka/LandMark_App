import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/favorites_api.dart';
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
