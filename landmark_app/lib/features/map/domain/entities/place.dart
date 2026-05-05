enum PlaceVisibility { public, private }

class Place {
  final String id;
  final String title;
  final double latitude;
  final double longitude;
  final String? description;
  final String? city;
  final String? country;
  final String? coverMediaId;
  final List<String> categorySlugs;
  final PlaceVisibility visibility;

  const Place({
    required this.id,
    required this.title,
    required this.latitude,
    required this.longitude,
    this.description,
    this.city,
    this.country,
    this.coverMediaId,
    this.categorySlugs = const [],
    this.visibility = PlaceVisibility.public,
  });

  bool get isPrivate => visibility == PlaceVisibility.private;

  factory Place.fromJson(Map<String, dynamic> j) => Place(
        id: j['id'] as String,
        title: j['title'] as String,
        latitude: (j['latitude'] as num).toDouble(),
        longitude: (j['longitude'] as num).toDouble(),
        description: j['description'] as String?,
        city: j['city'] as String?,
        country: j['country'] as String?,
        coverMediaId: j['cover_media_id'] as String?,
        categorySlugs:
            (j['category_slugs'] as List<dynamic>?)?.cast<String>() ?? [],
        visibility: (j['visibility'] as String?) == 'private'
            ? PlaceVisibility.private
            : PlaceVisibility.public,
      );
}

class PlaceCategory {
  final String id;
  final String slug;
  final String title;
  final String icon;
  final String color;

  const PlaceCategory({
    required this.id,
    required this.slug,
    required this.title,
    required this.icon,
    required this.color,
  });

  factory PlaceCategory.fromJson(Map<String, dynamic> j) => PlaceCategory(
        id: j['id'] as String,
        slug: j['slug'] as String,
        title: j['title'] as String,
        icon: j['icon'] as String? ?? '',
        color: j['color'] as String? ?? '#0E7C7B',
      );
}
