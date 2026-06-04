import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:landmark_app/app.dart';
import 'package:landmark_app/core/connectivity/connectivity_provider.dart';
import 'package:landmark_app/core/networking/api_client.dart';
import 'package:landmark_app/core/storage/secure_storage.dart';
import 'package:landmark_app/features/journal/data/local/draft_storage.dart';

void main() {
  testWidgets('App boots to splash without exception',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    final storage = SecureStorageImpl();
    final apiClient = ApiClient(
      baseUrl: 'http://localhost:8080',
      getAccessToken: () async => null,
      getRefreshToken: () async => null,
      saveTokens: (_, __) async {},
      clearTokens: () async {},
    );

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          secureStorageProvider.overrideWithValue(storage),
          apiClientProvider.overrideWithValue(apiClient),
          draftStorageProvider.overrideWithValue(DraftStorage(prefs)),
          // Prevent real DNS lookups and pending timers
          connectivityProvider.overrideWith((_) => Stream.value(true)),
        ],
        child: const App(),
      ),
    );

    // First frame — splash screen should render
    await tester.pump();
    expect(tester.takeException(), isNull);
    expect(find.byType(MaterialApp), findsOneWidget);

    // Advance past the 1500ms splash delay so its timer completes cleanly
    await tester.pump(const Duration(milliseconds: 1600));
    await tester.pumpAndSettle();
  });
}
