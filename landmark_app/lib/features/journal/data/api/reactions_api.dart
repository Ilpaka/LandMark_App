import 'package:dio/dio.dart';

class ReactionCount {
  final String emoji;
  final int count;
  const ReactionCount({required this.emoji, required this.count});

  factory ReactionCount.fromJson(Map<String, dynamic> j) =>
      ReactionCount(emoji: j['emoji'] as String, count: j['count'] as int);
}

class EntryReactions {
  final List<ReactionCount> counts;
  final List<String> mine;
  const EntryReactions({required this.counts, required this.mine});
}

class ReactionsApi {
  final Dio dio;
  const ReactionsApi(this.dio);

  Future<EntryReactions> list(String entryId) async {
    final r = await dio
        .get<Map<String, dynamic>>('/v1/journal/entries/$entryId/reactions');
    final body = r.data!;
    return EntryReactions(
      counts: (body['counts'] as List<dynamic>)
          .map((e) => ReactionCount.fromJson(e as Map<String, dynamic>))
          .toList(),
      mine: (body['mine'] as List<dynamic>).cast<String>(),
    );
  }

  Future<void> add(String entryId, String emoji) =>
      dio.post<void>('/v1/journal/entries/$entryId/reactions',
          data: {'emoji': emoji});

  Future<void> remove(String entryId, String emoji) =>
      dio.delete<void>('/v1/journal/entries/$entryId/reactions/$emoji');
}
