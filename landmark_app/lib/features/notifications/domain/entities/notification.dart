class AppNotification {
  final String id;
  final String type;
  final String title;
  final String body;
  final String? deepLink;
  final DateTime createdAt;
  final bool isRead;

  const AppNotification({
    required this.id,
    required this.type,
    required this.title,
    required this.body,
    this.deepLink,
    required this.createdAt,
    required this.isRead,
  });

  factory AppNotification.fromJson(Map<String, dynamic> j) => AppNotification(
        id: j['id'] as String,
        type: j['type'] as String? ?? 'info',
        title: j['title'] as String? ?? '',
        body: j['body'] as String? ?? '',
        deepLink: j['deep_link'] as String?,
        createdAt: DateTime.parse(j['created_at'] as String),
        isRead: j['read_at'] != null,
      );

  AppNotification copyWith({bool? isRead}) => AppNotification(
        id: id,
        type: type,
        title: title,
        body: body,
        deepLink: deepLink,
        createdAt: createdAt,
        isRead: isRead ?? this.isRead,
      );
}
