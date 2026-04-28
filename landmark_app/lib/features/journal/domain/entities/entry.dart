class JournalEntry {
  final String id;
  final String authorId;
  final String? tripId;
  final String? title;
  final String body;
  final String? mood;
  final int? rating;
  final DateTime occurredAt;
  final DateTime createdAt;
  final List<String> tags;

  const JournalEntry({
    required this.id,
    required this.authorId,
    this.tripId,
    this.title,
    required this.body,
    this.mood,
    this.rating,
    required this.occurredAt,
    required this.createdAt,
    this.tags = const [],
  });

  factory JournalEntry.fromJson(Map<String, dynamic> j) => JournalEntry(
    id: j['id'] as String,
    authorId: j['author_id'] as String,
    tripId: j['trip_id'] as String?,
    title: j['title'] as String?,
    body: j['body'] as String? ?? '',
    mood: j['mood'] as String?,
    rating: j['rating'] as int?,
    occurredAt: DateTime.parse(j['occurred_at'] as String),
    createdAt: DateTime.parse(j['created_at'] as String),
    tags: (j['tags'] as List<dynamic>?)?.cast<String>() ?? [],
  );
}
