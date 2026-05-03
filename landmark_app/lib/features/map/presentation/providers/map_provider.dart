import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/place.dart';
import '../../data/api/places_api.dart';
import '../../data/repositories/places_repository_impl.dart';
import '../../../../core/networking/api_client.dart';

final placesRepositoryProvider = Provider((ref) {
  final client = ref.read(apiClientProvider);
  return PlacesRepositoryImpl(PlacesApi(client.dio));
});

final categoriesProvider = FutureProvider<List<PlaceCategory>>((ref) {
  return ref.read(placesRepositoryProvider).getCategories();
});

// Параметры запроса мест
class PlacesQuery {
  final double? south, west, north, east;
  final String? categorySlug;
  final String? q;
  const PlacesQuery(
      {this.south,
      this.west,
      this.north,
      this.east,
      this.categorySlug,
      this.q});

  @override
  bool operator ==(Object other) =>
      other is PlacesQuery &&
      south == other.south &&
      west == other.west &&
      north == other.north &&
      east == other.east &&
      categorySlug == other.categorySlug &&
      q == other.q;
  @override
  int get hashCode => Object.hash(south, west, north, east, categorySlug, q);
}

final placesProvider =
    FutureProvider.family.autoDispose<List<Place>, PlacesQuery>((ref, query) {
  return ref.read(placesRepositoryProvider).listPlaces(
        southLat: query.south,
        westLng: query.west,
        northLat: query.north,
        eastLng: query.east,
        categorySlug: query.categorySlug,
        q: query.q,
      );
});

// Провайдер текущего выбранного фильтра категорий
final selectedCategoryProvider = StateProvider<String?>((ref) => null);

// Провайдер текущего bbox (обновляется при движении камеры)
final mapBBoxProvider = StateProvider<PlacesQuery?>((ref) => null);
