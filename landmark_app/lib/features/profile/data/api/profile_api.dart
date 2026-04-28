import 'package:dio/dio.dart';
import '../../domain/entities/profile.dart';

class ProfileApi {
  final Dio dio;
  ProfileApi(this.dio);

  Future<UserProfile?> getMe() async {
    try {
      final r = await dio.get('/v1/profile/me');
      return UserProfile.fromJson(r.data as Map<String, dynamic>);
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<UserProfile> updateMe({
    String? nickname,
    String? displayName,
    String? bio,
    String? city,
    String? country,
    String? avatarMediaId,
  }) async {
    final body = <String, dynamic>{};
    if (nickname != null) body['nickname'] = nickname;
    if (displayName != null) body['display_name'] = displayName;
    if (bio != null) body['bio'] = bio;
    if (city != null) body['city'] = city;
    if (country != null) body['country'] = country;
    if (avatarMediaId != null) body['avatar_media_id'] = avatarMediaId;
    final r = await dio.patch('/v1/profile/me', data: body);
    return UserProfile.fromJson(r.data as Map<String, dynamic>);
  }

  Future<PrivacySettings> getPrivacy() async {
    final r = await dio.get('/v1/profile/me/privacy');
    return PrivacySettings.fromJson(r.data as Map<String, dynamic>);
  }

  Future<PrivacySettings> updatePrivacy(Map<String, dynamic> patch) async {
    final r = await dio.patch('/v1/profile/me/privacy', data: patch);
    return PrivacySettings.fromJson(r.data as Map<String, dynamic>);
  }
}
