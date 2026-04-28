import '../entities/place.dart';

abstract class PlacesRepository {
  Future<List<Place>> listPlaces({
    double? southLat, double? westLng, double? northLat, double? eastLng,
    String? categorySlug, String? q, int limit = 100,
  });
  Future<Place> getPlace(String id);
  Future<List<PlaceCategory>> getCategories();
}
