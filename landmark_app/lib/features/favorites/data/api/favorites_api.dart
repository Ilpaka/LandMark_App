import 'package:dio/dio.dart';

class FavoritesApi {
  final Dio dio;
  FavoritesApi(this.dio);

  Future<bool> isFavorite(String targetType, String targetId) async {
    try {
      final r = await dio.get('/v1/favorites/$targetType/$targetId');
      return (r.data as Map<String, dynamic>)['is_favorite'] as bool? ?? false;
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return false;
      rethrow;
    }
  }

  Future<void> addFavorite(String targetType, String targetId) async {
    await dio.post('/v1/favorites', data: {
      'target_type': targetType,
      'target_id': targetId,
    });
  }

  Future<void> removeFavorite(String targetType, String targetId) async {
    await dio.delete('/v1/favorites/$targetType/$targetId');
  }

  Future<List<Map<String, dynamic>>> listFavorites({
    String? targetType,
    String? cursor,
  }) async {
    final params = <String, dynamic>{};
    if (targetType != null) params['target_type'] = targetType;
    if (cursor != null) params['cursor'] = cursor;
    final r = await dio.get('/v1/favorites', queryParameters: params);
    final data = r.data as Map<String, dynamic>;
    return (data['items'] as List<dynamic>?)
            ?.cast<Map<String, dynamic>>() ??
        [];
  }
}
