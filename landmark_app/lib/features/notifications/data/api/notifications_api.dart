import 'package:dio/dio.dart';
import '../../domain/entities/notification.dart';

class NotificationsPage {
  final List<AppNotification> items;
  final String? nextCursor;
  final int unreadCount;

  const NotificationsPage({
    required this.items,
    this.nextCursor,
    required this.unreadCount,
  });
}

class NotificationsApi {
  final Dio dio;
  const NotificationsApi(this.dio);

  Future<NotificationsPage> list({String? cursor, bool unreadOnly = false}) async {
    final r = await dio.get<Map<String, dynamic>>('/v1/notifications', queryParameters: {
      if (cursor != null) 'cursor': cursor,
      if (unreadOnly) 'unread': 'true',
    });
    final body = r.data!;
    return NotificationsPage(
      items: (body['items'] as List<dynamic>? ?? [])
          .map((e) => AppNotification.fromJson(e as Map<String, dynamic>))
          .toList(),
      nextCursor: body['next_cursor'] as String?,
      unreadCount: body['unread_count'] as int? ?? 0,
    );
  }

  Future<void> markRead(String id) =>
      dio.post<void>('/v1/notifications/$id/read');

  Future<void> markAllRead() =>
      dio.post<void>('/v1/notifications/read-all');
}
