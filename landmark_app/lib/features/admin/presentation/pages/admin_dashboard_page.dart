import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../providers/moderation_provider.dart';
import '../../data/api/moderation_api.dart';
import 'moderation_queue_page.dart';
import 'categories_page.dart';

class AdminDashboardPage extends ConsumerWidget {
  const AdminDashboardPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statsAsync = ref.watch(moderationStatsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Панель администратора'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              ref.invalidate(moderationStatsProvider);
              ref.invalidate(moderationQueueProvider);
            },
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Stats cards
          statsAsync.when(
            loading: () => const _StatsCardsSkeleton(),
            error: (e, _) => _ErrorCard(error: e, onRetry: () => ref.invalidate(moderationStatsProvider)),
            data: (stats) => _StatsGrid(stats: stats),
          ),
          const SizedBox(height: 24),
          // Actions section
          const Text(
            'УПРАВЛЕНИЕ',
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: AppColors.textSecondary,
              letterSpacing: 0.8,
            ),
          ),
          const SizedBox(height: 12),
          _ActionCard(
            icon: Icons.pending_actions_outlined,
            title: 'Очередь модерации',
            subtitle: statsAsync.maybeWhen(
              data: (s) => s.pending > 0 ? '${s.pending} ожидает проверки' : 'Нет заявок',
              orElse: () => 'Заявки на проверку мест',
            ),
            badgeCount: statsAsync.maybeWhen(
              data: (s) => s.pending,
              orElse: () => 0,
            ),
            onTap: () => Navigator.push<void>(
              context,
              MaterialPageRoute(builder: (_) => const ModerationQueuePage()),
            ),
          ),
          const SizedBox(height: 8),
          _ActionCard(
            icon: Icons.category_outlined,
            title: 'Категории мест',
            subtitle: 'Управление категориями POI',
            onTap: () => Navigator.push<void>(
              context,
              MaterialPageRoute(builder: (_) => const CategoriesPage()),
            ),
          ),
          const SizedBox(height: 8),
          _ActionCard(
            icon: Icons.people_outline,
            title: 'Пользователи',
            subtitle: 'Управление аккаунтами',
            onTap: () => ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Раздел в разработке')),
            ),
          ),
        ],
      ),
    );
  }
}

class _StatsGrid extends StatelessWidget {
  final ModerationStats stats;
  const _StatsGrid({required this.stats});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'СТАТИСТИКА МОДЕРАЦИИ',
          style: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.w600,
            color: AppColors.textSecondary,
            letterSpacing: 0.8,
          ),
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(child: _StatCard(
              label: 'Ожидает',
              value: stats.pending,
              color: Colors.orange,
              icon: Icons.hourglass_empty,
            )),
            const SizedBox(width: 10),
            Expanded(child: _StatCard(
              label: 'Одобрено',
              value: stats.approved,
              color: Colors.green,
              icon: Icons.check_circle_outline,
            )),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Expanded(child: _StatCard(
              label: 'Отклонено',
              value: stats.rejected,
              color: Colors.red,
              icon: Icons.cancel_outlined,
            )),
            const SizedBox(width: 10),
            Expanded(child: _StatCard(
              label: 'Всего',
              value: stats.total,
              color: AppColors.primary,
              icon: Icons.list_alt_outlined,
            )),
          ],
        ),
      ],
    );
  }
}

class _StatCard extends StatelessWidget {
  final String label;
  final int value;
  final Color color;
  final IconData icon;
  const _StatCard({required this.label, required this.value, required this.color, required this.icon});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.2)),
      ),
      child: Row(
        children: [
          Icon(icon, color: color, size: 28),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '$value',
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.w700,
                  color: color,
                ),
              ),
              Text(
                label,
                style: const TextStyle(fontSize: 12, color: AppColors.textSecondary),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _ActionCard extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final int badgeCount;
  final VoidCallback onTap;
  const _ActionCard({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
    this.badgeCount = 0,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      color: AppColors.surface,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Row(
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: AppColors.primary.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(icon, color: AppColors.primary, size: 22),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(title,
                        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 15)),
                    const SizedBox(height: 2),
                    Text(subtitle,
                        style: const TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                  ],
                ),
              ),
              if (badgeCount > 0)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: Colors.orange,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    '$badgeCount',
                    style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.w700),
                  ),
                )
              else
                const Icon(Icons.chevron_right, color: AppColors.textSecondary),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatsCardsSkeleton extends StatelessWidget {
  const _StatsCardsSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Row(children: [
          Expanded(child: Container(height: 80, decoration: BoxDecoration(
            color: AppColors.surface, borderRadius: BorderRadius.circular(12)))),
          const SizedBox(width: 10),
          Expanded(child: Container(height: 80, decoration: BoxDecoration(
            color: AppColors.surface, borderRadius: BorderRadius.circular(12)))),
        ]),
        const SizedBox(height: 10),
        Row(children: [
          Expanded(child: Container(height: 80, decoration: BoxDecoration(
            color: AppColors.surface, borderRadius: BorderRadius.circular(12)))),
          const SizedBox(width: 10),
          Expanded(child: Container(height: 80, decoration: BoxDecoration(
            color: AppColors.surface, borderRadius: BorderRadius.circular(12)))),
        ]),
      ],
    );
  }
}

class _ErrorCard extends StatelessWidget {
  final Object error;
  final VoidCallback onRetry;
  const _ErrorCard({required this.error, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return Card(
      color: Colors.red.shade50,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            const Icon(Icons.error_outline, color: Colors.red),
            const SizedBox(width: 12),
            Expanded(child: Text('Ошибка загрузки: $error',
                style: const TextStyle(fontSize: 13))),
            TextButton(onPressed: onRetry, child: const Text('Повтор')),
          ],
        ),
      ),
    );
  }
}
