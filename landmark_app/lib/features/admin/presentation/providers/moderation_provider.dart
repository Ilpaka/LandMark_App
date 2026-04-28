import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/moderation_api.dart';
import '../../../../core/networking/api_client.dart';

final moderationApiProvider = Provider((ref) {
  return ModerationApi(ref.read(apiClientProvider).dio);
});

final moderationStatsProvider = FutureProvider<ModerationStats>(
  (ref) => ref.read(moderationApiProvider).getStats(),
);

final moderationQueueProvider =
    AsyncNotifierProvider<ModerationQueueNotifier, List<ModerationQueueItem>>(
  ModerationQueueNotifier.new,
);

class ModerationQueueNotifier extends AsyncNotifier<List<ModerationQueueItem>> {
  @override
  Future<List<ModerationQueueItem>> build() async {
    return ref.read(moderationApiProvider).listQueue();
  }

  Future<void> approve(String itemId) async {
    await ref.read(moderationApiProvider).approve(itemId);
    ref.invalidateSelf();
  }

  Future<void> reject(String itemId, String note) async {
    await ref.read(moderationApiProvider).reject(itemId, note);
    ref.invalidateSelf();
  }
}
