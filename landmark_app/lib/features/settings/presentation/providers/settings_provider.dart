import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/settings_api.dart';
import '../../../../core/networking/api_client.dart';

final settingsApiProvider = Provider((ref) {
  return SettingsApi(ref.read(apiClientProvider).dio);
});

final notifPrefsProvider =
    AsyncNotifierProvider<NotifPrefsNotifier, NotificationPreferences>(
  NotifPrefsNotifier.new,
);

class NotifPrefsNotifier extends AsyncNotifier<NotificationPreferences> {
  @override
  Future<NotificationPreferences> build() =>
      ref.read(settingsApiProvider).getNotifPrefs();

  Future<void> toggle(String key, bool value) async {
    final current = state.valueOrNull;
    if (current == null) return;
    final updated =
        await ref.read(settingsApiProvider).updateNotifPrefs({key: value});
    state = AsyncValue.data(updated);
  }
}
