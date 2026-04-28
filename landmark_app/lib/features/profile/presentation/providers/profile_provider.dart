import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/profile_api.dart';
import '../../domain/entities/profile.dart';
import '../../../../core/networking/api_client.dart';

final profileApiProvider = Provider((ref) {
  return ProfileApi(ref.read(apiClientProvider).dio);
});

final myProfileProvider = AsyncNotifierProvider<MyProfileNotifier, UserProfile?>(
  MyProfileNotifier.new,
);

class MyProfileNotifier extends AsyncNotifier<UserProfile?> {
  @override
  Future<UserProfile?> build() => ref.read(profileApiProvider).getMe();

  Future<void> save({
    String? nickname,
    String? displayName,
    String? bio,
    String? city,
    String? country,
    String? avatarMediaId,
  }) async {
    final updated = await ref.read(profileApiProvider).updateMe(
      nickname: nickname,
      displayName: displayName,
      bio: bio,
      city: city,
      country: country,
      avatarMediaId: avatarMediaId,
    );
    state = AsyncValue.data(updated);
  }
}

final privacyProvider = AsyncNotifierProvider<PrivacyNotifier, PrivacySettings>(
  PrivacyNotifier.new,
);

class PrivacyNotifier extends AsyncNotifier<PrivacySettings> {
  @override
  Future<PrivacySettings> build() async {
    return ref.read(profileApiProvider).getPrivacy();
  }

  Future<void> save(Map<String, dynamic> patch) async {
    final updated = await ref.read(profileApiProvider).updatePrivacy(patch);
    state = AsyncValue.data(updated);
  }
}
