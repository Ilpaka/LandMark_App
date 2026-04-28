import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/design/theme.dart';
import 'core/routing/router.dart';
import 'core/widgets/offline_banner.dart';
import 'features/journal/data/sync/draft_sync_service.dart';

class App extends ConsumerWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // Initialise draft sync listener (kept alive for App lifetime)
    ref.watch(draftSyncProvider);
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      title: 'Wanderlog',
      theme: AppTheme.light(),
      routerConfig: router,
      debugShowCheckedModeBanner: false,
      builder: (context, child) => Column(
        children: [
          const OfflineBanner(),
          Expanded(child: child ?? const SizedBox.shrink()),
        ],
      ),
    );
  }
}
