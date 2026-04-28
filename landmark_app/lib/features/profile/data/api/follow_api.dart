import 'package:dio/dio.dart';
import '../../domain/entities/profile.dart';

class FollowStats {
  final int followersCount;
  final int followingCount;
  final bool isFollowing;

  const FollowStats({
    required this.followersCount,
    required this.followingCount,
    required this.isFollowing,
  });

  factory FollowStats.fromJson(Map<String, dynamic> j) => FollowStats(
        followersCount: (j['followers_count'] as num).toInt(),
        followingCount: (j['following_count'] as num).toInt(),
        isFollowing: j['is_following'] as bool? ?? false,
      );

  FollowStats copyWith({bool? isFollowing, int? followersCount, int? followingCount}) => FollowStats(
        followersCount: followersCount ?? this.followersCount,
        followingCount: followingCount ?? this.followingCount,
        isFollowing: isFollowing ?? this.isFollowing,
      );
}

class FollowApi {
  final Dio dio;
  FollowApi(this.dio);

  Future<FollowStats> getStats(String userId) async {
    final r = await dio.get('/v1/profile/$userId/stats');
    return FollowStats.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> follow(String userId) => dio.post('/v1/profile/$userId/follow');

  Future<void> unfollow(String userId) => dio.delete('/v1/profile/$userId/follow');

  Future<List<UserProfile>> listFollowers(String userId) async {
    final r = await dio.get('/v1/profile/$userId/followers');
    return (r.data as List).map((e) => UserProfile.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<UserProfile>> listFollowing(String userId) async {
    final r = await dio.get('/v1/profile/$userId/following');
    return (r.data as List).map((e) => UserProfile.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<UserProfile> getPublicProfile(String userId) async {
    final r = await dio.get('/v1/profile/$userId');
    return UserProfile.fromJson(r.data as Map<String, dynamic>);
  }
}
