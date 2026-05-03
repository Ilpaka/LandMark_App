import 'package:dio/dio.dart';
import '../../domain/entities/place.dart';

class PlacesApi {
  final Dio dio;
  PlacesApi(this.dio);

  Future<List<Place>> listPlaces({
    double? southLat,
    double? westLng,
    double? northLat,
    double? eastLng,
    String? categorySlug,
    String? q,
    int limit = 100,
  }) async {
    final params = <String, dynamic>{'limit': limit};
    if (southLat != null &&
        westLng != null &&
        northLat != null &&
        eastLng != null) {
      params['bbox'] = '$southLat,$westLng,$northLat,$eastLng';
    }
    if (categorySlug != null) params['category'] = categorySlug;
    if (q != null && q.isNotEmpty) params['q'] = q;
    final r = await dio.get('/v1/places', queryParameters: params);
    final data = r.data as Map<String, dynamic>;
    final list = data['places'] as List<dynamic>? ?? [];
    return list.map((e) => Place.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Place> getPlace(String id) async {
    final r = await dio.get('/v1/places/$id');
    return Place.fromJson(r.data as Map<String, dynamic>);
  }

  Future<List<PlaceCategory>> getCategories() async {
    final r = await dio.get('/v1/places/categories');
    final data = r.data as Map<String, dynamic>;
    final list = data['categories'] as List<dynamic>? ?? [];
    return list
        .map((e) => PlaceCategory.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
