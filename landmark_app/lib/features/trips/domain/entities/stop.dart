class TripStop {
  final String id;
  final String tripId;
  final String? placeId;
  final String title;
  final double latitude;
  final double longitude;
  final String? note;
  final int order;
  final DateTime? visitedAt;
  final DateTime createdAt;

  const TripStop({
    required this.id,
    required this.tripId,
    this.placeId,
    required this.title,
    required this.latitude,
    required this.longitude,
    this.note,
    required this.order,
    this.visitedAt,
    required this.createdAt,
  });

  factory TripStop.fromJson(Map<String, dynamic> j) => TripStop(
        id: j['id'] as String,
        tripId: j['trip_id'] as String,
        placeId: j['place_id'] as String?,
        title: j['title'] as String,
        latitude: (j['latitude'] as num).toDouble(),
        longitude: (j['longitude'] as num).toDouble(),
        note: j['note'] as String?,
        order: j['order'] as int? ?? 0,
        visitedAt: j['visited_at'] != null
            ? DateTime.parse(j['visited_at'] as String)
            : null,
        createdAt: DateTime.parse(j['created_at'] as String),
      );
}
