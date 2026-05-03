import 'dart:io';
import 'package:dio/dio.dart';

class MediaApi {
  final Dio dio;
  const MediaApi(this.dio);

  /// Возвращает список mediaId для записи журнала.
  Future<List<String>> getEntryMedia(String entryId) async {
    final r = await dio
        .get<Map<String, dynamic>>('/v1/journal/entries/$entryId/media');
    final ids = r.data!['media_ids'] as List<dynamic>? ?? [];
    return ids.cast<String>();
  }

  /// Возвращает presigned URL для скачивания медиа-файла.
  Future<String> getMediaUrl(String mediaId) async {
    final r = await dio.get<Map<String, dynamic>>('/v1/media/$mediaId/url');
    return r.data!['url'] as String;
  }

  /// Прикрепляет ранее загруженный mediaId к записи журнала.
  Future<void> attachMedia(
      String entryId, String mediaId, int sortOrder) async {
    await dio.post('/v1/journal/entries/$entryId/media', data: {
      'media_id': mediaId,
      'sort_order': sortOrder,
    });
  }

  /// Загружает файл и возвращает mediaId.
  Future<String> uploadFile(File file, {String purpose = 'photo'}) async {
    final mime = _contentType(file.path);
    final size = await file.length();

    // 1. Request presigned upload URL
    final init =
        await dio.post<Map<String, dynamic>>('/v1/media/uploads', data: {
      'kind': purpose, // 'photo' or 'avatar'
      'mime': mime,
      'size_bytes': size,
    });
    final uploadId = init.data!['upload_id'] as String;
    final presignedURL = init.data!['presigned_url'] as String;

    // 2. PUT file to presigned URL (no auth header for S3/MinIO presigned)
    final uploadDio = Dio();
    await uploadDio.put(
      presignedURL,
      data: file.openRead(),
      options: Options(
        headers: {
          Headers.contentLengthHeader: size,
          'Content-Type': mime,
        },
        sendTimeout: const Duration(minutes: 2),
        receiveTimeout: const Duration(minutes: 2),
      ),
    );

    // 3. Finalize → returns Media with its own id
    final fin = await dio.post<Map<String, dynamic>>(
        '/v1/media/uploads/$uploadId/finalize');
    return fin.data!['id'] as String;
  }

  String _contentType(String path) {
    final ext = path.split('.').last.toLowerCase();
    return switch (ext) {
      'jpg' || 'jpeg' => 'image/jpeg',
      'png' => 'image/png',
      'webp' => 'image/webp',
      _ => 'application/octet-stream',
    };
  }
}
