import 'dart:async';
import 'dart:io';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final connectivityProvider = StreamProvider<bool>((ref) {
  final controller = StreamController<bool>();

  Future<bool> check() async {
    try {
      final result = await InternetAddress.lookup('google.com')
          .timeout(const Duration(seconds: 3));
      return result.isNotEmpty && result.first.rawAddress.isNotEmpty;
    } catch (_) {
      return false;
    }
  }

  // Poll every 5 seconds
  final timer = Timer.periodic(const Duration(seconds: 5), (_) async {
    if (!controller.isClosed) {
      controller.add(await check());
    }
  });

  // Initial check
  check().then((v) {
    if (!controller.isClosed) controller.add(v);
  });

  ref.onDispose(() {
    timer.cancel();
    controller.close();
  });
  return controller.stream;
});

final isOfflineProvider = Provider<bool>((ref) {
  return ref.watch(connectivityProvider).when(
        data: (online) => !online,
        loading: () => false,
        error: (_, __) => true,
      );
});
