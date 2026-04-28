class UserProfile {
  final String userId;
  final String nickname;
  final String displayName;
  final String? bio;
  final String? avatarMediaId;
  final String? city;
  final String? country;
  final DateTime createdAt;
  final DateTime updatedAt;

  const UserProfile({
    required this.userId,
    required this.nickname,
    required this.displayName,
    this.bio,
    this.avatarMediaId,
    this.city,
    this.country,
    required this.createdAt,
    required this.updatedAt,
  });

  factory UserProfile.fromJson(Map<String, dynamic> j) => UserProfile(
        userId: j['user_id'] as String,
        nickname: j['nickname'] as String,
        displayName: j['display_name'] as String,
        bio: j['bio'] as String?,
        avatarMediaId: j['avatar_media_id'] as String?,
        city: j['city'] as String?,
        country: j['country'] as String?,
        createdAt: DateTime.parse(j['created_at'] as String),
        updatedAt: DateTime.parse(j['updated_at'] as String),
      );
}

class PrivacySettings {
  final String profileVisibility;
  final String tripsVisibility;
  final String journalVisibility;
  final bool analyticsEnabled;

  const PrivacySettings({
    required this.profileVisibility,
    required this.tripsVisibility,
    required this.journalVisibility,
    required this.analyticsEnabled,
  });

  factory PrivacySettings.fromJson(Map<String, dynamic> j) => PrivacySettings(
        profileVisibility: j['profile_visibility'] as String? ?? 'public',
        tripsVisibility: j['trips_visibility'] as String? ?? 'private',
        journalVisibility: j['journal_visibility'] as String? ?? 'private',
        analyticsEnabled: j['analytics_enabled'] as bool? ?? true,
      );

  PrivacySettings copyWith({
    String? profileVisibility,
    String? tripsVisibility,
    String? journalVisibility,
    bool? analyticsEnabled,
  }) =>
      PrivacySettings(
        profileVisibility: profileVisibility ?? this.profileVisibility,
        tripsVisibility: tripsVisibility ?? this.tripsVisibility,
        journalVisibility: journalVisibility ?? this.journalVisibility,
        analyticsEnabled: analyticsEnabled ?? this.analyticsEnabled,
      );
}
