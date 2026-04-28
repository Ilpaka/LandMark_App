import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../data/api/follow_api.dart';
import '../../domain/entities/profile.dart';
import '../providers/follow_provider.dart';
import '../widgets/user_avatar.dart';
import 'followers_page.dart';

class PublicProfilePage extends ConsumerWidget {
  final String userId;
  const PublicProfilePage({super.key, required this.userId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(publicProfileProvider(userId));
    final AsyncValue<FollowStats> statsAsync = ref.watch(followStatsProvider(userId));

    return Scaffold(
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: profileAsync.when(
        loading: () => const Center(child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (profile) => _ProfileBody(
          profile: profile,
          statsAsync: statsAsync,
          userId: userId,
          onToggleFollow: () => ref.read(followStatsProvider(userId).notifier).toggle(),
        ),
      ),
    );
  }
}

class _ProfileBody extends StatelessWidget {
  final UserProfile profile;
  final AsyncValue<FollowStats> statsAsync;
  final String userId;
  final VoidCallback onToggleFollow;

  const _ProfileBody({
    required this.profile,
    required this.statsAsync,
    required this.userId,
    required this.onToggleFollow,
  });


  @override
  Widget build(BuildContext context) {
    return ListView(
      children: [
        Container(
          padding: const EdgeInsets.all(24),
          color: AppColors.surface,
          child: Column(
            children: [
              UserAvatar(
                displayName: profile.displayName,
                mediaId: profile.avatarMediaId,
                radius: 44,
              ),
              const SizedBox(height: 12),
              Text(
                profile.displayName,
                style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w700, color: AppColors.textPrimary),
              ),
              const SizedBox(height: 4),
              Text('@${profile.nickname}',
                  style: const TextStyle(fontSize: 15, color: AppColors.textSecondary)),
              if (profile.bio != null && profile.bio!.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  profile.bio!,
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 14, color: AppColors.textSecondary),
                ),
              ],
              if (profile.city != null || profile.country != null) ...[
                const SizedBox(height: 6),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.location_on_outlined, size: 14, color: AppColors.textSecondary),
                    const SizedBox(width: 4),
                    Text(
                      [profile.city, profile.country].where((s) => s != null).join(', '),
                      style: const TextStyle(fontSize: 13, color: AppColors.textSecondary),
                    ),
                  ],
                ),
              ],
              const SizedBox(height: 16),
              statsAsync.when(
                loading: () => const SizedBox(height: 48, child: Center(child: CircularProgressIndicator(strokeWidth: 2))),
                error: (_, __) => const SizedBox.shrink(),
                data: (stats) => _StatsRow(stats: stats, userId: userId),
              ),
              const SizedBox(height: 16),
              statsAsync.when(
                loading: () => const SizedBox.shrink(),
                error: (_, __) => const SizedBox.shrink(),
                data: (stats) => SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: stats.isFollowing ? AppColors.surface : AppColors.primary,
                      foregroundColor: stats.isFollowing ? AppColors.primary : Colors.white,
                      side: stats.isFollowing
                          ? const BorderSide(color: AppColors.primary)
                          : BorderSide.none,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      padding: const EdgeInsets.symmetric(vertical: 12),
                    ),
                    onPressed: onToggleFollow,
                    child: Text(stats.isFollowing ? 'Отписаться' : 'Подписаться',
                        style: const TextStyle(fontWeight: FontWeight.w600)),
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _StatsRow extends StatelessWidget {
  final FollowStats stats;
  final String userId;
  const _StatsRow({required this.stats, required this.userId});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        _StatItem(
          count: stats.followersCount,
          label: 'подписчиков',
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => FollowersPage(userId: userId, showFollowers: true)),
          ),
        ),
        const SizedBox(width: 40),
        _StatItem(
          count: stats.followingCount,
          label: 'подписок',
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => FollowersPage(userId: userId, showFollowers: false)),
          ),
        ),
      ],
    );
  }
}

class _StatItem extends StatelessWidget {
  final int count;
  final String label;
  final VoidCallback onTap;
  const _StatItem({required this.count, required this.label, required this.onTap});

  @override
  Widget build(BuildContext context) => GestureDetector(
    onTap: onTap,
    child: Column(
      children: [
        Text('$count',
            style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: AppColors.textPrimary)),
        const SizedBox(height: 2),
        Text(label, style: const TextStyle(fontSize: 12, color: AppColors.textSecondary)),
      ],
    ),
  );
}
