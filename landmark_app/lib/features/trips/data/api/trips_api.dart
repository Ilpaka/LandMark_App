import 'package:dio/dio.dart';
import '../../domain/entities/trip.dart';

class TripsApi {
  final Dio dio;
  TripsApi(this.dio);

  Future<List<Trip>> listTrips({String? status, int limit = 20}) async {
    final r = await dio.get('/v1/trips', queryParameters: {
      if (status != null) 'status': status,
      'limit': limit,
    });
    final data = r.data as Map<String, dynamic>;
    final list = data['trips'] as List<dynamic>? ?? [];
    return list.map((e) => Trip.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Trip> createTrip(String title) async {
    final r = await dio.post('/v1/trips', data: {'title': title});
    return Trip.fromJson(r.data as Map<String, dynamic>);
  }

  Future<Trip> getTrip(String id) async {
    final r = await dio.get('/v1/trips/$id');
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

  Future<List<Map<String, double>>> getRoute(String tripId) async {
    final r = await dio.get('/v1/trips/$tripId/route');
    final data = r.data as Map<String, dynamic>;
    final coords = data['coordinates'] as List<dynamic>? ?? [];
    return coords.map((c) {
      final m = c as Map<String, dynamic>;
      return {'lat': (m['lat'] as num).toDouble(), 'lng': (m['lng'] as num).toDouble()};
    }).toList();
  }
}
