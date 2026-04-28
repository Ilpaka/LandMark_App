import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../domain/entities/profile.dart';
import '../providers/profile_provider.dart';
import '../../../auth/presentation/providers/auth_provider.dart' show authProvider, AuthStateAuthenticated; // ignore: unused_shown_name
import 'edit_profile_page.dart';
import '../../../settings/presentation/pages/settings_page.dart';

class ProfilePage extends ConsumerWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(myProfileProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Профиль'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            onPressed: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const SettingsPage()),
            ),
          ),
        ],
      ),
      body: profileAsync.when(
        loading: () => const Center(child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => _buildEmpty(context, ref),
        data: (profile) => profile == null
            ? _buildEmpty(context, ref)
            : _buildProfile(context, ref, profile),
      ),
    );
  }

  Widget _buildEmpty(BuildContext context, WidgetRef ref) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.person_outline, size: 72, color: AppColors.textSecondary),
          const SizedBox(height: 16),
          const Text('Профиль не настроен', style: TextStyle(fontSize: 18, color: AppColors.textSecondary)),
          const SizedBox(height: 24),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primary, foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            ),
            onPressed: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const EditProfilePage()),
            ),
            child: const Text('Создать профиль'),
          ),
        ],
      ),
    );
  }

  Widget _buildProfile(BuildContext context, WidgetRef ref, UserProfile profile) {
    return ListView(
      children: [
        // Header
        Container(
          padding: const EdgeInsets.all(24),
          color: AppColors.surface,
          child: Column(
            children: [
              CircleAvatar(
                radius: 44,
                backgroundColor: AppColors.primary.withValues(alpha: 0.15),
                child: Text(
                  (profile.displayName.isNotEmpty ? profile.displayName[0] : '?').toUpperCase(),
                  style: const TextStyle(fontSize: 36, color: AppColors.primary, fontWeight: FontWeight.w600),
                ),
              ),
              const SizedBox(height: 12),
              Text(profile.displayName,
                  style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w700, color: AppColors.textPrimary)),
              const SizedBox(height: 4),
              Text('@${profile.nickname}',
                  style: const TextStyle(fontSize: 15, color: AppColors.textSecondary)),
              if (profile.bio != null && profile.bio!.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(profile.bio!,
                    textAlign: TextAlign.center,
                    style: const TextStyle(fontSize: 14, color: AppColors.textSecondary)),
              ],
              if (profile.city != null || profile.country != null) ...[
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.location_on_outlined, size: 16, color: AppColors.textSecondary),
                    const SizedBox(width: 4),
                    Text(
                      [profile.city, profile.country].where((s) => s != null).join(', '),
                      style: const TextStyle(fontSize: 13, color: AppColors.textSecondary),
                    ),
                  ],
                ),
              ],
              const SizedBox(height: 16),
              OutlinedButton.icon(
                icon: const Icon(Icons.edit_outlined, size: 16),
                label: const Text('Редактировать'),
                style: OutlinedButton.styleFrom(
                  foregroundColor: AppColors.primary,
                  side: const BorderSide(color: AppColors.primary),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                ),
                onPressed: () => Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => EditProfilePage(profile: profile)),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 8),
        // Sign out
        ListTile(
          leading: const Icon(Icons.logout, color: Colors.red),
          title: const Text('Выйти', style: TextStyle(color: Colors.red)),
          onTap: () => _confirmSignOut(context, ref),
        ),
      ],
    );
  }

  void _confirmSignOut(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Выход'),
        content: const Text('Вы уверены, что хотите выйти?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Отмена')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red, foregroundColor: Colors.white),
            onPressed: () {
              Navigator.pop(ctx);
              ref.read(authProvider.notifier).signOut();
            },
            child: const Text('Выйти'),
          ),
        ],
      ),
    );
  }

}
