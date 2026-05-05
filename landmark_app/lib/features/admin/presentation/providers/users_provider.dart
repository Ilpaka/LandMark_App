import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/networking/api_client.dart';
import '../../data/api/users_api.dart';

final usersApiProvider = Provider<UsersApi>((ref) {
  return UsersApi(ref.read(apiClientProvider).dio);
});

final usersListProvider =
    AsyncNotifierProviderFamily<UsersListNotifier, List<AdminUser>, String?>(
  UsersListNotifier.new,
);

class UsersListNotifier
    extends FamilyAsyncNotifier<List<AdminUser>, String?> {
  @override
  Future<List<AdminUser>> build(String? query) async {
    return ref.read(usersApiProvider).list(query: query);
  }

  Future<void> refreshSelf() async {
    state = const AsyncValue.loading();
    state = await AsyncValue.guard(
        () => ref.read(usersApiProvider).list(query: arg));
  }

  Future<void> block(String userId, String reason) async {
    await ref.read(usersApiProvider).block(userId, reason);
    await refreshSelf();
  }

  Future<void> unblock(String userId) async {
    await ref.read(usersApiProvider).unblock(userId);
    await refreshSelf();
  }

  Future<void> forceLogout(String userId) async {
    await ref.read(usersApiProvider).forceLogout(userId);
  }
}

final auditProvider =
    FutureProviderFamily<List<AdminAuditEvent>, String>((ref, userId) {
  return ref.read(usersApiProvider).audit(userId);
});
