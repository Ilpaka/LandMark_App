import 'package:dio/dio.dart';

class MediaApi {
  final Dio dio;
  const MediaApi(this.dio);

  /// Возвращает список mediaId для записи журнала.
  Future<List<String>> getEntryMedia(String entryId) async {
    final r = await dio.get<Map<String, dynamic>>(
        '/v1/journal/entries/$entryId/media');
    final ids = r.data!['media_ids'] as List<dynamic>? ?? [];
    return ids.cast<String>();
  }

  /// Возвращает presigned URL для скачивания медиа-файла.
  Future<String> getMediaUrl(String mediaId) async {
    final r = await dio.get<Map<String, dynamic>>('/v1/media/$mediaId/url');
    return r.data!['url'] as String;
  }
}
