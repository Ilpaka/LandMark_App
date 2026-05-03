import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/notifications_api.dart';
import '../../domain/entities/notification.dart';
import '../../../../core/networking/api_client.dart';

final notificationsApiProvider = Provider<NotificationsApi>(
  (ref) => NotificationsApi(ref.read(apiClientProvider).dio),
);

// ---------------------------------------------------------------------------
// Состояние ленты уведомлений
// ---------------------------------------------------------------------------

class NotificationsState {
  final List<AppNotification> items;
  final String? nextCursor;
  final int unreadCount;
  final bool isLoading;
  final String? error;

  const NotificationsState({
    this.items = const [],
    this.nextCursor,
    this.unreadCount = 0,
    this.isLoading = false,
    this.error,
  });

  NotificationsState copyWith({
    List<AppNotification>? items,
    String? nextCursor,
    int? unreadCount,
    bool? isLoading,
    String? error,
  }) =>
      NotificationsState(
        items: items ?? this.items,
        nextCursor: nextCursor,
        unreadCount: unreadCount ?? this.unreadCount,
        isLoading: isLoading ?? this.isLoading,
        error: error,
      );
}

final notificationsFeedProvider =
    AsyncNotifierProvider<NotificationsFeedNotifier, NotificationsState>(
  NotificationsFeedNotifier.new,
);

class NotificationsFeedNotifier extends AsyncNotifier<NotificationsState> {
  @override
  Future<NotificationsState> build() => _fetch();

  Future<NotificationsState> _fetch({String? cursor}) async {
    final api = ref.read(notificationsApiProvider);
    final page = await api.list(cursor: cursor);
    return NotificationsState(
      items: page.items,
      nextCursor: page.nextCursor,
      unreadCount: page.unreadCount,
    );
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(_fetch);
  }

  Future<void> markRead(String id) async {
    final current = state.valueOrNull;
    if (current == null) return;

    // Оптимистичное обновление
    final updated = current.items.map((n) {
      return n.id == id ? n.copyWith(isRead: true) : n;
    }).toList();
    final newUnread = updated.where((n) => !n.isRead).length;
    state = AsyncData(current.copyWith(items: updated, unreadCount: newUnread));

    try {
      await ref.read(notificationsApiProvider).markRead(id);
    } catch (_) {
      state = AsyncData(current);
    }
  }

  Future<void> markAllRead() async {
    final current = state.valueOrNull;
    if (current == null) return;

    final updated = current.items.map((n) => n.copyWith(isRead: true)).toList();
    state = AsyncData(current.copyWith(items: updated, unreadCount: 0));

    try {
      await ref.read(notificationsApiProvider).markAllRead();
    } catch (_) {
      state = AsyncData(current);
    }
  }
}

// Провайдер только количества непрочитанных (для бейджа)
final unreadCountProvider = Provider<int>((ref) {
  final feed = ref.watch(notificationsFeedProvider);
  return feed.valueOrNull?.unreadCount ?? 0;
});
