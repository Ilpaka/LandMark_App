import 'package:dio/dio.dart';
import '../../domain/entities/stop.dart';

class StopsApi {
  final Dio dio;
  StopsApi(this.dio);

  Future<List<TripStop>> listStops(String tripId) async {
    final r = await dio.get('/v1/trips/$tripId/stops');
    final data = r.data as Map<String, dynamic>;
    return (data['stops'] as List<dynamic>? ?? [])
        .map((e) => TripStop.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<TripStop> addStop(
    String tripId, {
    required String title,
    required double latitude,
    required double longitude,
    String? note,
    String? placeId,
  }) async {
    final r = await dio.post('/v1/trips/$tripId/stops', data: {
      'title': title,
      'latitude': latitude,
      'longitude': longitude,
      if (note != null) 'note': note,
      if (placeId != null) 'place_id': placeId,
    });
    return TripStop.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> markVisited(String tripId, String stopId) async {
    await dio.post('/v1/trips/$tripId/stops/$stopId/visit');
  }
}
