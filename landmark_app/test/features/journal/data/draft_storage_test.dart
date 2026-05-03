import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:landmark_app/features/journal/data/local/draft_storage.dart';

JournalDraft _draft(String id, {String title = 'd', String body = 'b'}) =>
    JournalDraft(
      id: id,
      tripId: 't-1',
      title: title,
      body: body,
      createdAt: DateTime.utc(2026, 3, 1, 12),
    );

void main() {
  setUp(() => SharedPreferences.setMockInitialValues(<String, Object>{}));

  group('DraftStorage', () {
    test('returns empty list when no drafts have been saved', () async {
      final prefs = await SharedPreferences.getInstance();
      final storage = DraftStorage(prefs);

      expect(storage.getDrafts(), isEmpty);
    });

    test('saveDraft appends a new draft and persists it', () async {
      final prefs = await SharedPreferences.getInstance();
      final storage = DraftStorage(prefs);

      await storage.saveDraft(_draft('1', title: 'Hello'));
      await storage.saveDraft(_draft('2', title: 'World'));

      final drafts = storage.getDrafts();
      expect(drafts.map((d) => d.id), <String>['1', '2']);
      expect(drafts[0].title, 'Hello');
      expect(drafts[1].title, 'World');
    });

    test('saveDraft updates an existing draft in place by id', () async {
      final prefs = await SharedPreferences.getInstance();
      final storage = DraftStorage(prefs);

      await storage.saveDraft(_draft('1', title: 'First'));
      await storage.saveDraft(_draft('1', title: 'Updated'));

      final drafts = storage.getDrafts();
      expect(drafts, hasLength(1));
      expect(drafts.first.title, 'Updated');
    });

    test('removeDraft removes the matching draft only', () async {
      final prefs = await SharedPreferences.getInstance();
      final storage = DraftStorage(prefs);
      await storage.saveDraft(_draft('1'));
      await storage.saveDraft(_draft('2'));

      await storage.removeDraft('1');

      expect(storage.getDrafts().map((d) => d.id), <String>['2']);
    });

    test('mood field round-trips through JSON serialization', () async {
      final prefs = await SharedPreferences.getInstance();
      final storage = DraftStorage(prefs);

      await storage.saveDraft(JournalDraft(
        id: '1',
        tripId: 't-1',
        title: 'with mood',
        body: 'b',
        mood: 'reflective',
        createdAt: DateTime.utc(2026, 3, 1),
      ));

      // Re-read through a fresh storage instance to bypass any in-memory state.
      final replay = DraftStorage(await SharedPreferences.getInstance());
      expect(replay.getDrafts().single.mood, 'reflective');
    });
  });
}
