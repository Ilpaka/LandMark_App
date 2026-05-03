import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/entry.dart';
import '../../data/api/journal_api.dart';
import '../../../../core/networking/api_client.dart';

final journalApiProvider =
    Provider((ref) => JournalApi(ref.read(apiClientProvider).dio));

final journalFeedProvider =
    AsyncNotifierProvider.autoDispose<JournalFeedNotifier, List<JournalEntry>>(
  JournalFeedNotifier.new,
);

class JournalFeedNotifier extends AutoDisposeAsyncNotifier<List<JournalEntry>> {
  String? _nextCursor;
  bool _hasMore = true;

  @override
  Future<List<JournalEntry>> build() async {
    _nextCursor = null;
    _hasMore = true;
    final result = await ref.read(journalApiProvider).listEntries();
    _nextCursor = result.nextCursor;
    _hasMore = result.nextCursor != null;
    return result.entries;
  }

  bool get hasMore => _hasMore;

  Future<void> loadMore() async {
    if (!_hasMore || _nextCursor == null) return;
    final current = state.valueOrNull ?? [];
    final result =
        await ref.read(journalApiProvider).listEntries(cursor: _nextCursor);
    _nextCursor = result.nextCursor;
    _hasMore = result.nextCursor != null;
    state = AsyncData([...current, ...result.entries]);
  }
}
