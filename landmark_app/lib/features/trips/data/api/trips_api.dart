import 'package:dio/dio.dart';
import '../../domain/entities/trip.dart';

class TripsApi {
  final Dio dio;
  TripsApi(this.dio);

  Future<({List<Trip> trips, String? nextCursor})> listTrips({
    String? status,
    String? cursor,
    int limit = 20,
  }) async {
    final r = await dio.get('/v1/trips', queryParameters: {
      if (status != null) 'status': status,
      if (cursor != null) 'cursor': cursor,
      'limit': limit,
    });
    final data = r.data as Map<String, dynamic>;
    final list = data['trips'] as List<dynamic>? ?? [];
    return (
      trips: list.map((e) => Trip.fromJson(e as Map<String, dynamic>)).toList(),
      nextCursor: data['next_cursor'] as String?,
    );
  }

  Future<Trip> createTrip(String title) async {
    final r = await dio.post('/v1/trips', data: {'title': title});
    return Trip.fromJson(r.data as Map<String, dynamic>);
  }

  Future<Trip> getTrip(String id) async {
    final r = await dio.get('/v1/trips/$id');
    return Trip.fromJson(r.data as Map<String, dynamic>);
  }

  Future<Trip> updateTrip(String id,
      {String? title, String? subtitle, String? status}) async {
    final body = <String, dynamic>{};
    if (title != null) body['title'] = title;
    if (subtitle != null) body['subtitle'] = subtitle;
    if (status != null) body['status'] = status;
    final r = await dio.patch('/v1/trips/$id', data: body);
    return Trip.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> deleteTrip(String id) => dio.delete('/v1/trips/$id');

  Future<List<Trip>> listPublicTrips({String? cursor, int limit = 20}) async {
    final r = await dio.get('/v1/trips/public', queryParameters: {
      if (cursor != null) 'cursor': cursor,
      'limit': limit,
    });
    final data = r.data as Map<String, dynamic>;
    final list = data['trips'] as List<dynamic>? ?? [];
    return list.map((e) => Trip.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<String>> optimizeRoute(String tripId) async {
    final r = await dio.post('/v1/trips/$tripId/optimize');
    final data = r.data as Map<String, dynamic>;
    return (data['ordered_stop_ids'] as List<dynamic>? ?? [])
        .map((e) => e as String)
        .toList();
  }

  Future<List<Map<String, double>>> getRoute(String tripId) async {
    final r = await dio.get('/v1/trips/$tripId/route');
    final data = r.data as Map<String, dynamic>;
    final coords = data['coordinates'] as List<dynamic>? ?? [];
    return coords.map((c) {
      final m = c as Map<String, dynamic>;
      return {
        'lat': (m['lat'] as num).toDouble(),
        'lng': (m['lng'] as num).toDouble()
      };
    }).toList();
  }
}
