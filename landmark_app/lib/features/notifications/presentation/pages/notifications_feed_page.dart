import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/widgets/skeleton.dart';
import '../../domain/entities/notification.dart';
import '../providers/notifications_provider.dart';

class NotificationsFeedPage extends ConsumerWidget {
  const NotificationsFeedPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final feedAsync = ref.watch(notificationsFeedProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Уведомления'),
        actions: [
          feedAsync.valueOrNull?.unreadCount != null &&
                  feedAsync.valueOrNull!.unreadCount > 0
              ? TextButton(
                  onPressed: () =>
                      ref.read(notificationsFeedProvider.notifier).markAllRead(),
                  child: const Text('Прочитать все',
                      style: TextStyle(color: AppColors.primary)),
                )
              : const SizedBox.shrink(),
        ],
      ),
      body: feedAsync.when(
        loading: () => const SkeletonListView(),
        error: (e, _) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline,
                  size: 48, color: AppColors.textSecondary),
              const SizedBox(height: 12),
              const Text('Не удалось загрузить уведомления'),
              const SizedBox(height: 12),
              TextButton(
                onPressed: () =>
                    ref.read(notificationsFeedProvider.notifier).refresh(),
                child: const Text('Повторить'),
              ),
            ],
          ),
        ),
        data: (state) => state.items.isEmpty
            ? const _EmptyNotifications()
            : RefreshIndicator(
                color: AppColors.primary,
                onRefresh: () =>
                    ref.read(notificationsFeedProvider.notifier).refresh(),
                child: ListView.separated(
                  itemCount: state.items.length,
                  separatorBuilder: (_, __) =>
                      const Divider(height: 1, indent: 72),
                  itemBuilder: (_, i) => _NotificationTile(
                    notification: state.items[i],
                    onTap: () => ref
                        .read(notificationsFeedProvider.notifier)
                        .markRead(state.items[i].id),
                  ),
                ),
              ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Плитка уведомления
// ---------------------------------------------------------------------------

class _NotificationTile extends StatelessWidget {
  final AppNotification notification;
  final VoidCallback onTap;

  const _NotificationTile({
    required this.notification,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final fmt = DateFormat('d MMM, HH:mm', 'ru_RU');

    return InkWell(
      onTap: onTap,
      child: Container(
        color: notification.isRead
            ? null
            : AppColors.primary.withValues(alpha: 0.05),
        padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Иконка типа
            Container(
              width: 44,
              height: 44,
              margin: const EdgeInsets.only(right: AppSpacing.md),
              decoration: BoxDecoration(
                color: _iconBg(notification.type),
                shape: BoxShape.circle,
              ),
              child: Icon(
                _iconData(notification.type),
                color: _iconColor(notification.type),
                size: 22,
              ),
            ),
            // Текст
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          notification.title,
                          style: TextStyle(
                            fontWeight: notification.isRead
                                ? FontWeight.w400
                                : FontWeight.w600,
                            fontSize: 14,
                            color: AppColors.textPrimary,
                          ),
                        ),
                      ),
                      Text(
                        fmt.format(notification.createdAt.toLocal()),
                        style: AppTypography.caption,
                      ),
                    ],
                  ),
                  const SizedBox(height: 2),
                  Text(
                    notification.body,
                    style: AppTypography.bodySmall,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
            // Индикатор непрочитанного
            if (!notification.isRead)
              Container(
                width: 8,
                height: 8,
                margin: const EdgeInsets.only(left: 8, top: 4),
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  shape: BoxShape.circle,
                ),
              ),
          ],
        ),
      ),
    );
  }

  IconData _iconData(String type) => switch (type) {
        'moderation_approved' => Icons.check_circle_outline,
        'moderation_rejected' => Icons.cancel_outlined,
        'trip_reminder' => Icons.luggage_outlined,
        'sync_complete' => Icons.sync,
        _ => Icons.notifications_outlined,
      };

  Color _iconBg(String type) => switch (type) {
        'moderation_approved' => Colors.green.withValues(alpha: 0.15),
        'moderation_rejected' => Colors.red.withValues(alpha: 0.12),
        'trip_reminder' => AppColors.accent.withValues(alpha: 0.15),
        'sync_complete' => Colors.blue.withValues(alpha: 0.12),
        _ => AppColors.primary.withValues(alpha: 0.1),
      };

  Color _iconColor(String type) => switch (type) {
        'moderation_approved' => Colors.green,
        'moderation_rejected' => Colors.red,
        'trip_reminder' => AppColors.accent,
        'sync_complete' => Colors.blue,
        _ => AppColors.primary,
      };
}

// ---------------------------------------------------------------------------
// Пустое состояние (NT-02)
// ---------------------------------------------------------------------------

class _EmptyNotifications extends StatelessWidget {
  const _EmptyNotifications();

  @override
  Widget build(BuildContext context) => const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.notifications_none_outlined,
                size: 72, color: AppColors.textSecondary),
            SizedBox(height: 16),
            Text(
              'Нет уведомлений',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
            ),
            SizedBox(height: 8),
            Text(
              'Здесь будут появляться важные\nуведомления о ваших поездках',
              style: TextStyle(color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      );
}
