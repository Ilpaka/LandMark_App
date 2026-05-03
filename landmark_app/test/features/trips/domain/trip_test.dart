import 'package:flutter_test/flutter_test.dart';
import 'package:landmark_app/features/trips/domain/entities/trip.dart';

void main() {
  group('Trip.fromJson', () {
    test('parses required fields and uses defaults for counts', () {
      final trip = Trip.fromJson(<String, dynamic>{
        'id': 'trip-42',
        'title': 'Northern Lights',
        'created_at': '2026-01-10T08:00:00Z',
        'updated_at': '2026-01-12T10:30:00Z',
      });

      expect(trip.id, 'trip-42');
      expect(trip.title, 'Northern Lights');
      expect(trip.subtitle, isNull);
      expect(trip.status, 'draft');
      expect(trip.stopsCount, 0);
      expect(trip.entriesCount, 0);
      expect(trip.photosCount, 0);
      expect(trip.createdAt.toUtc(), DateTime.utc(2026, 1, 10, 8));
    });

    test('parses optional subtitle and counts when present', () {
      final trip = Trip.fromJson(<String, dynamic>{
        'id': 'trip-1',
        'title': 'Iceland',
        'subtitle': 'ring road',
        'status': 'active',
        'created_at': '2026-02-01T09:00:00Z',
        'updated_at': '2026-02-15T18:00:00Z',
        'stops_count': 8,
        'entries_count': 3,
        'photos_count': 27,
      });

      expect(trip.subtitle, 'ring road');
      expect(trip.status, 'active');
      expect(trip.stopsCount, 8);
      expect(trip.entriesCount, 3);
      expect(trip.photosCount, 27);
    });

    test('throws FormatException when timestamps are malformed', () {
      expect(
        () => Trip.fromJson(<String, dynamic>{
          'id': 'trip-x',
          'title': 'Broken',
          'created_at': 'not-a-date',
          'updated_at': '2026-02-15T18:00:00Z',
        }),
        throwsFormatException,
      );
    });
  });
}
