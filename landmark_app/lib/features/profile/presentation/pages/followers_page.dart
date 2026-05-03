import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../domain/entities/profile.dart';
import '../providers/follow_provider.dart';
import '../widgets/user_avatar.dart';
import 'public_profile_page.dart';

class FollowersPage extends ConsumerWidget {
  final String userId;
  final bool showFollowers;
  const FollowersPage(
      {super.key, required this.userId, required this.showFollowers});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final listAsync = showFollowers
        ? ref.watch(followersProvider(userId))
        : ref.watch(followingProvider(userId));

    return Scaffold(
      appBar: AppBar(
        title: Text(showFollowers ? 'Подписчики' : 'Подписки'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: listAsync.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (profiles) => profiles.isEmpty
            ? Center(
                child: Text(
                  showFollowers ? 'Нет подписчиков' : 'Нет подписок',
                  style: const TextStyle(
                      color: AppColors.textSecondary, fontSize: 16),
                ),
              )
            : ListView.builder(
                itemCount: profiles.length,
                itemBuilder: (_, i) => _ProfileTile(profile: profiles[i]),
              ),
      ),
    );
  }
}

class _ProfileTile extends StatelessWidget {
  final UserProfile profile;
  const _ProfileTile({required this.profile});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: UserAvatar(
        displayName: profile.displayName,
        mediaId: profile.avatarMediaId,
        radius: 22,
      ),
      title: Text(profile.displayName,
          style: const TextStyle(
              fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
      subtitle: Text('@${profile.nickname}',
          style: const TextStyle(color: AppColors.textSecondary, fontSize: 13)),
      onTap: () => Navigator.push(
        context,
        MaterialPageRoute(
            builder: (_) => PublicProfilePage(userId: profile.userId)),
      ),
    );
  }
}
