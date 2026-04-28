import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../profile/presentation/providers/profile_provider.dart';
import '../providers/settings_provider.dart';

class SettingsPage extends ConsumerWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Настройки'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: ListView(
        children: [
          _sectionHeader('Уведомления'),
          _NotifPrefsSection(),
          const Divider(height: 1),
          _sectionHeader('Приватность'),
          _PrivacySection(),
        ],
      ),
    );
  }

  Widget _sectionHeader(String title) => Padding(
        padding: const EdgeInsets.fromLTRB(16, 20, 16, 8),
        child: Text(title,
            style: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: AppColors.textSecondary,
                letterSpacing: 0.5)),
      );
}

class _NotifPrefsSection extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final prefsAsync = ref.watch(notifPrefsProvider);

    return prefsAsync.when(
      loading: () => const Padding(
        padding: EdgeInsets.all(16),
        child: Center(child: CircularProgressIndicator(color: AppColors.primary, strokeWidth: 2)),
      ),
      error: (_, __) => const ListTile(
        leading: Icon(Icons.error_outline, color: Colors.red),
        title: Text('Не удалось загрузить настройки'),
      ),
      data: (prefs) => Column(
        children: [
          _toggle(
            ref, 'Напоминания о поездках', prefs.pushTripReminders,
            'push_trip_reminders',
            subtitle: 'Push-уведомления',
          ),
          _toggle(
            ref, 'Результат модерации', prefs.pushModerationResult,
            'push_moderation_result',
            subtitle: 'Push-уведомления',
          ),
          _toggle(
            ref, 'Результат модерации (email)', prefs.emailModerationResult,
            'email_moderation_result',
            subtitle: 'Email-уведомления',
          ),
          _toggle(
            ref, 'Маркетинговые (email)', prefs.emailMarketing,
            'email_marketing',
            subtitle: 'Email-уведомления',
          ),
        ],
      ),
    );
  }

  Widget _toggle(WidgetRef ref, String title, bool value, String key, {String? subtitle}) {
    return SwitchListTile(
      value: value,
      title: Text(title),
      subtitle: subtitle != null ? Text(subtitle, style: const TextStyle(fontSize: 12)) : null,
      activeColor: AppColors.primary,
      onChanged: (v) => ref.read(notifPrefsProvider.notifier).toggle(key, v),
    );
  }
}

class _PrivacySection extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final privacyAsync = ref.watch(privacyProvider);

    return privacyAsync.when(
      loading: () => const Padding(
        padding: EdgeInsets.all(16),
        child: Center(child: CircularProgressIndicator(color: AppColors.primary, strokeWidth: 2)),
      ),
      error: (_, __) => const ListTile(
        leading: Icon(Icons.error_outline, color: Colors.red),
        title: Text('Не удалось загрузить настройки'),
      ),
      data: (privacy) => Column(
        children: [
          _dropdownTile(
            ref, context,
            'Профиль',
            privacy.profileVisibility,
            ['public', 'private'],
            ['Публичный', 'Приватный'],
            (v) => ref.read(privacyProvider.notifier).save({'profile_visibility': v}),
          ),
          _dropdownTile(
            ref, context,
            'Поездки',
            privacy.tripsVisibility,
            ['public', 'friends', 'private'],
            ['Публичные', 'Только друзья', 'Приватные'],
            (v) => ref.read(privacyProvider.notifier).save({'trips_visibility': v}),
          ),
          _dropdownTile(
            ref, context,
            'Журнал',
            privacy.journalVisibility,
            ['friends', 'private'],
            ['Только друзья', 'Приватный'],
            (v) => ref.read(privacyProvider.notifier).save({'journal_visibility': v}),
          ),
          SwitchListTile(
            value: privacy.analyticsEnabled,
            title: const Text('Аналитика'),
            subtitle: const Text('Разрешить сбор аналитики', style: TextStyle(fontSize: 12)),
            activeColor: AppColors.primary,
            onChanged: (v) => ref.read(privacyProvider.notifier).save({'analytics_enabled': v}),
          ),
        ],
      ),
    );
  }

  Widget _dropdownTile(
    WidgetRef ref,
    BuildContext context,
    String title,
    String currentValue,
    List<String> values,
    List<String> labels,
    ValueChanged<String> onChanged,
  ) {
    return ListTile(
      title: Text(title),
      trailing: DropdownButton<String>(
        value: currentValue,
        underline: const SizedBox(),
        style: const TextStyle(color: AppColors.textPrimary, fontSize: 14),
        items: List.generate(
          values.length,
          (i) => DropdownMenuItem(value: values[i], child: Text(labels[i])),
        ),
        onChanged: (v) { if (v != null) onChanged(v); },
      ),
    );
  }
}
