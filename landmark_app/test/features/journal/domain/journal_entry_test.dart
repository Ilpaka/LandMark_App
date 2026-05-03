import 'package:flutter_test/flutter_test.dart';
import 'package:landmark_app/features/journal/domain/entities/entry.dart';

void main() {
  group('JournalEntry.fromJson', () {
    test('parses full payload including optional fields and tags list', () {
      final entry = JournalEntry.fromJson(<String, dynamic>{
        'id': 'e-1',
        'author_id': 'u-1',
        'trip_id': 't-1',
        'title': 'Day one',
        'body': 'Arrived in Reykjavik.',
        'mood': 'happy',
        'rating': 5,
        'occurred_at': '2026-02-01T12:00:00Z',
        'created_at': '2026-02-01T13:00:00Z',
        'tags': <String>['travel', 'iceland'],
      });

      expect(entry.id, 'e-1');
      expect(entry.authorId, 'u-1');
      expect(entry.tripId, 't-1');
      expect(entry.title, 'Day one');
      expect(entry.body, 'Arrived in Reykjavik.');
      expect(entry.mood, 'happy');
      expect(entry.rating, 5);
      expect(entry.tags, <String>['travel', 'iceland']);
    });

    test('uses safe defaults when optional fields are missing', () {
      final entry = JournalEntry.fromJson(<String, dynamic>{
        'id': 'e-2',
        'author_id': 'u-2',
        'occurred_at': '2026-02-02T08:00:00Z',
        'created_at': '2026-02-02T09:00:00Z',
      });

      expect(entry.tripId, isNull);
      expect(entry.title, isNull);
      expect(entry.body, '');
      expect(entry.mood, isNull);
      expect(entry.rating, isNull);
      expect(entry.tags, isEmpty);
    });
  });
}
