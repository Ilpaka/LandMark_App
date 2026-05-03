class Trip {
  final String id;
  final String title;
  final String? subtitle;
  final String status;
  final DateTime createdAt;
  final DateTime updatedAt;
  final int stopsCount;
  final int entriesCount;
  final int photosCount;

  const Trip({
    required this.id,
    required this.title,
    this.subtitle,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
    this.stopsCount = 0,
    this.entriesCount = 0,
    this.photosCount = 0,
  });

  factory Trip.fromJson(Map<String, dynamic> j) => Trip(
        id: j['id'] as String,
        title: j['title'] as String,
        subtitle: j['subtitle'] as String?,
        status: j['status'] as String? ?? 'draft',
        createdAt: DateTime.parse(j['created_at'] as String),
        updatedAt: DateTime.parse(j['updated_at'] as String),
        stopsCount: j['stops_count'] as int? ?? 0,
        entriesCount: j['entries_count'] as int? ?? 0,
        photosCount: j['photos_count'] as int? ?? 0,
      );
}
