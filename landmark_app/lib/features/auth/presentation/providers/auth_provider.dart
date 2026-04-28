import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../data/api/auth_api.dart';
import '../../data/repositories/auth_repository_impl.dart';
import '../../../../core/networking/api_client.dart';
import '../../../../core/storage/secure_storage.dart';

// Repository provider
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final client = ref.read(apiClientProvider);
  final storage = ref.read(secureStorageProvider);
  return AuthRepositoryImpl(AuthApi(client.dio), storage);
});

// Auth state
sealed class AuthState {}
class AuthStateUnknown extends AuthState {}
class AuthStateUnauthenticated extends AuthState {}
class AuthStateAuthenticated extends AuthState {
  final User user;
  AuthStateAuthenticated(this.user);
}

class AuthNotifier extends StateNotifier<AuthState> {
  final AuthRepository _repo;

  AuthNotifier(this._repo) : super(AuthStateUnknown()) {
    _bootstrap();
  }

  Future<void> _bootstrap() async {
    final user = await _repo.tryRestoreSession();
    state = user == null ? AuthStateUnauthenticated() : AuthStateAuthenticated(user);
  }

  Future<String> register(String name, String email, String password) async {
    return _repo.register(name: name, email: email, password: password);
  }

  Future<void> verifyEmail(String verificationId, String code) async {
    await _repo.verifyEmail(verificationId: verificationId, code: code);
  }

  Future<void> login(String email, String password) async {
    final user = await _repo.login(email: email, password: password);
    state = AuthStateAuthenticated(user);
  }

  Future<void> forgotPassword(String email) => _repo.forgotPassword(email: email);

  Future<void> signOut() async {
    await _repo.signOut();
    state = AuthStateUnauthenticated();
  }

  Future<void> deleteAccount() async {
    await _repo.deleteAccount();
    state = AuthStateUnauthenticated();
  }
}

// A ChangeNotifier that GoRouter can use as refreshListenable.
// It notifies whenever authProvider state changes.
class AuthRouterNotifier extends ChangeNotifier {
  AuthRouterNotifier(Ref ref) {
    ref.listen<AuthState>(authProvider, (_, __) => notifyListeners());
  }
}

final authRouterNotifierProvider = Provider<AuthRouterNotifier>((ref) {
  return AuthRouterNotifier(ref);
});

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref.read(authRepositoryProvider));
});
