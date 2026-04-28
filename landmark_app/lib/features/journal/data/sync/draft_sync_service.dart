import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../api/journal_api.dart';
import '../local/draft_storage.dart';
import '../../../../core/connectivity/connectivity_provider.dart';
import '../../presentation/providers/journal_provider.dart';

final draftSyncProvider = Provider<DraftSyncService>((ref) {
  final service = DraftSyncService(
    ref.read(draftStorageProvider),
    ref.read(journalApiProvider),
  );

  // Flush drafts whenever connectivity is restored
  ref.listen<AsyncValue<bool>>(connectivityProvider, (prev, next) {
    final wasOffline = prev?.valueOrNull == false || prev == null;
    final isOnline = next.valueOrNull == true;
    if (wasOffline && isOnline) {
      service.flush();
    }
  });

  // Also attempt flush on first load if online
  ref.read(connectivityProvider).whenData((online) {
    if (online) service.flush();
  });

  return service;
});

class DraftSyncService {
  final DraftStorage _storage;
  final JournalApi _api;
  bool _syncing = false;

  DraftSyncService(this._storage, this._api);

  Future<void> flush() async {
    if (_syncing) return;
    _syncing = true;
    try {
      final drafts = _storage.getDrafts();
      for (final draft in drafts) {
        try {
          await _api.createEntry(
            tripId: draft.tripId,
            title: draft.title.isEmpty ? null : draft.title,
            body: draft.body,
            mood: draft.mood,
          );
          await _storage.removeDraft(draft.id);
        } catch (_) {
          // Keep draft for next sync attempt
        }
      }
    } finally {
      _syncing = false;
    }
  }
}
