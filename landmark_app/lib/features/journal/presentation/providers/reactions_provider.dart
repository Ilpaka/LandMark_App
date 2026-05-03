import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/networking/api_client.dart';
import '../../data/api/reactions_api.dart';

final reactionsApiProvider = Provider<ReactionsApi>(
  (ref) => ReactionsApi(ref.read(apiClientProvider).dio),
);

final reactionsProvider =
    AsyncNotifierProvider.family<ReactionsNotifier, EntryReactions, String>(
  ReactionsNotifier.new,
);

class ReactionsNotifier extends FamilyAsyncNotifier<EntryReactions, String> {
  @override
  Future<EntryReactions> build(String entryId) =>
      ref.read(reactionsApiProvider).list(entryId);

  Future<void> toggle(String emoji) async {
    final current = state.valueOrNull;
    if (current == null) return;

    final api = ref.read(reactionsApiProvider);
    final isActive = current.mine.contains(emoji);

    // Optimistic update
    final newMine = isActive
        ? (List<String>.from(current.mine)..remove(emoji))
        : [...current.mine, emoji];

    final newCounts = List<ReactionCount>.from(current.counts);
    final idx = newCounts.indexWhere((c) => c.emoji == emoji);
    if (isActive) {
      if (idx >= 0) {
        if (newCounts[idx].count <= 1) {
          newCounts.removeAt(idx);
        } else {
          newCounts[idx] =
              ReactionCount(emoji: emoji, count: newCounts[idx].count - 1);
        }
      }
    } else {
      if (idx >= 0) {
        newCounts[idx] =
            ReactionCount(emoji: emoji, count: newCounts[idx].count + 1);
      } else {
        newCounts.add(ReactionCount(emoji: emoji, count: 1));
      }
    }

    state = AsyncData(EntryReactions(counts: newCounts, mine: newMine));

    try {
      if (isActive) {
        await api.remove(arg, emoji);
      } else {
        await api.add(arg, emoji);
      }
    } catch (_) {
      // Rollback on failure
      state = AsyncData(current);
    }
  }
}
