import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../core/networking/api_client.dart';
import '../core/storage/secure_storage.dart';
import '../bootstrap/env.dart';
import '../features/journal/data/local/draft_storage.dart';

Future<List<Override>> buildProviderOverrides() async {
  final secureStorage = SecureStorageImpl();
  final apiClient = ApiClient(
    baseUrl: AppEnv.apiBaseUrl,
    getAccessToken: () => secureStorage.read('auth.access'),
    getRefreshToken: () => secureStorage.read('auth.refresh'),
    saveTokens: (access, refresh) async {
      await secureStorage.write('auth.access', access);
      await secureStorage.write('auth.refresh', refresh);
    },
    clearTokens: () async {
      await secureStorage.delete('auth.access');
      await secureStorage.delete('auth.refresh');
    },
  );

  final prefs = await SharedPreferences.getInstance();
  final draftStorage = DraftStorage(prefs);

  return [
    secureStorageProvider.overrideWithValue(secureStorage),
    apiClientProvider.overrideWithValue(apiClient),
    draftStorageProvider.overrideWithValue(draftStorage),
  ];
}
