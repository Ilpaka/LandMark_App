import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/entry.dart';
import '../../data/api/journal_api.dart';
import '../../../../core/networking/api_client.dart';

final journalApiProvider = Provider((ref) => JournalApi(ref.read(apiClientProvider).dio));

final journalFeedProvider = FutureProvider.autoDispose<List<JournalEntry>>((ref) {
  return ref.read(journalApiProvider).listEntries();
});
