import 'dart:io';
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

  /// Прикрепляет ранее загруженный mediaId к записи журнала.
  Future<void> attachMedia(String entryId, String mediaId, int sortOrder) async {
    await dio.post('/v1/journal/entries/$entryId/media', data: {
      'media_id': mediaId,
      'sort_order': sortOrder,
    });
  }

  /// Загружает файл и возвращает mediaId.
  Future<String> uploadFile(File file, {String purpose = 'avatar'}) async {
    // 1. Request presigned upload URL
    final init = await dio.post<Map<String, dynamic>>('/v1/media/upload', data: {
      'filename': file.path.split('/').last,
      'content_type': _contentType(file.path),
      'purpose': purpose,
    });
    final mediaId = init.data!['media_id'] as String;
    final uploadUrl = init.data!['upload_url'] as String;

    // 2. PUT file to presigned URL (no auth header needed for S3/MinIO presigned)
    final uploadDio = Dio();
    await uploadDio.put(
      uploadUrl,
      data: file.openRead(),
      options: Options(
        headers: {
          Headers.contentLengthHeader: await file.length(),
          'Content-Type': _contentType(file.path),
        },
        sendTimeout: const Duration(minutes: 2),
        receiveTimeout: const Duration(minutes: 2),
      ),
    );

    // 3. Finalize
    await dio.post('/v1/media/$mediaId/finalize');
    return mediaId;
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
