import 'package:dio/dio.dart';

class ModerationQueueItem {
  final String id;
  final String targetType;
  final String targetId;
  final String submittedBy;
  final String status;
  final String? note;
  final DateTime createdAt;

  const ModerationQueueItem({
    required this.id,
    required this.targetType,
    required this.targetId,
    required this.submittedBy,
    required this.status,
    this.note,
    required this.createdAt,
  });

  factory ModerationQueueItem.fromJson(Map<String, dynamic> j) => ModerationQueueItem(
        id: j['id'] as String,
        targetType: j['target_type'] as String,
        targetId: j['target_id'] as String,
        submittedBy: j['submitted_by'] as String,
        status: j['status'] as String,
        note: j['note'] as String?,
        createdAt: DateTime.parse(j['created_at'] as String),
      );
}

class ModerationApi {
  final Dio dio;
  ModerationApi(this.dio);

  Future<List<ModerationQueueItem>> listQueue({String? cursor}) async {
    final params = <String, dynamic>{};
    if (cursor != null) params['cursor'] = cursor;
    final r = await dio.get('/v1/moderation/queue', queryParameters: params);
    final data = r.data as Map<String, dynamic>;
    return (data['items'] as List<dynamic>? ?? [])
        .map((e) => ModerationQueueItem.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> approve(String itemId) async {
    await dio.post('/v1/moderation/queue/$itemId/approve');
  }

  Future<void> reject(String itemId, String note) async {
    await dio.post('/v1/moderation/queue/$itemId/reject', data: {'note': note});
  }
}
