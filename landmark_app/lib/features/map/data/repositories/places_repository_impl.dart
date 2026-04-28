import '../../domain/entities/place.dart';
import '../../domain/repositories/places_repository.dart';
import '../api/places_api.dart';

class PlacesRepositoryImpl implements PlacesRepository {
  final PlacesApi _api;
  PlacesRepositoryImpl(this._api);

  @override
  Future<List<Place>> listPlaces({
    double? southLat, double? westLng, double? northLat, double? eastLng,
    String? categorySlug, String? q, int limit = 100,
  }) => _api.listPlaces(
    southLat: southLat, westLng: westLng, northLat: northLat, eastLng: eastLng,
    categorySlug: categorySlug, q: q, limit: limit,
  );

  @override
  Future<Place> getPlace(String id) => _api.getPlace(id);

  @override
  Future<List<PlaceCategory>> getCategories() => _api.getCategories();
}
