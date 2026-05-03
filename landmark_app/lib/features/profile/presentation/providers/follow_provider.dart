import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/follow_api.dart';
import '../../domain/entities/profile.dart';
import '../../../../core/networking/api_client.dart';

final followApiProvider =
    Provider((ref) => FollowApi(ref.read(apiClientProvider).dio));

// Follow stats for a given userId (autoDispose so it refreshes on re-entry)
final followStatsProvider =
    AsyncNotifierProviderFamily<FollowStatsNotifier, FollowStats, String>(
  FollowStatsNotifier.new,
);

class FollowStatsNotifier extends FamilyAsyncNotifier<FollowStats, String> {
  @override
  Future<FollowStats> build(String arg) =>
      ref.read(followApiProvider).getStats(arg);

  Future<void> toggle() async {
    final current = state.valueOrNull;
    if (current == null) return;

    // Optimistic update
    final optimistic = current.copyWith(
      isFollowing: !current.isFollowing,
      followersCount: current.isFollowing
          ? current.followersCount - 1
          : current.followersCount + 1,
    );
    state = AsyncValue.data(optimistic);

    try {
      if (current.isFollowing) {
        await ref.read(followApiProvider).unfollow(arg);
      } else {
        await ref.read(followApiProvider).follow(arg);
      }
    } catch (_) {
      // Rollback
      state = AsyncValue.data(current);
    }
  }
}

// Public profile for a given userId
final publicProfileProvider = FutureProvider.autoDispose
    .family<UserProfile, String>(
        (ref, userId) => ref.read(followApiProvider).getPublicProfile(userId));

// Followers list
final followersProvider = FutureProvider.autoDispose
    .family<List<UserProfile>, String>(
        (ref, userId) => ref.read(followApiProvider).listFollowers(userId));

// Following list
final followingProvider = FutureProvider.autoDispose
    .family<List<UserProfile>, String>(
        (ref, userId) => ref.read(followApiProvider).listFollowing(userId));
