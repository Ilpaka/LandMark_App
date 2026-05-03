import 'package:dio/dio.dart';
import '../../domain/entities/entry.dart';

class JournalApi {
  final Dio dio;
  JournalApi(this.dio);

  Future<({List<JournalEntry> entries, String? nextCursor})> listEntries({
    String? tripId,
    String? cursor,
    int limit = 20,
  }) async {
    final r = await dio.get('/v1/journal/entries', queryParameters: {
      if (tripId != null) 'trip_id': tripId,
      if (cursor != null) 'cursor': cursor,
      'limit': limit,
    });
    final data = r.data as Map<String, dynamic>;
    final list = data['entries'] as List<dynamic>? ?? [];
    return (
      entries: list
          .map((e) => JournalEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
      nextCursor: data['next_cursor'] as String?,
    );
  }

  Future<JournalEntry> getEntry(String id) async {
    final r = await dio.get('/v1/journal/entries/$id');
    return JournalEntry.fromJson(r.data as Map<String, dynamic>);
  }

  Future<JournalEntry> createEntry({
    String? tripId,
    String? title,
    required String body,
    String? mood,
  }) async {
    final r = await dio.post('/v1/journal/entries', data: {
      if (tripId != null) 'trip_id': tripId,
      if (title != null) 'title': title,
      'body': body,
      if (mood != null) 'mood': mood,
    });
    return JournalEntry.fromJson(r.data as Map<String, dynamic>);
  }
}
