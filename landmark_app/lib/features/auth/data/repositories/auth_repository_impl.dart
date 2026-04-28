import 'package:dio/dio.dart';
import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';
import '../api/auth_api.dart';
import '../../../../core/storage/secure_storage.dart';

class AuthRepositoryImpl implements AuthRepository {
  final AuthApi _api;
  final SecureStorageImpl _storage;

  AuthRepositoryImpl(this._api, this._storage);

  @override
  Future<User?> tryRestoreSession() async {
    final token = await _storage.read('auth.access');
    if (token == null || token.isEmpty) return null;
    try {
      final data = await _api.me();
      return User.fromJson(data);
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        await _storage.deleteAll();
      }
      return null;
    }
  }

  @override
  Future<String> register({required String name, required String email, required String password}) async {
    final data = await _api.register(name, email, password);
    return data['verification_id'] as String;
  }

  @override
  Future<void> verifyEmail({required String verificationId, required String code}) async {
    await _api.verifyEmail(verificationId, code);
  }

  @override
  Future<User> login({required String email, required String password}) async {
    final data = await _api.login(email, password);
    await _storage.write('auth.access', data['access_token'] as String);
    await _storage.write('auth.refresh', data['refresh_token'] as String);
    return User.fromJson(data['user'] as Map<String, dynamic>? ?? data);
  }

  @override
  Future<void> forgotPassword({required String email}) => _api.forgot(email);

  @override
  Future<User> me() async {
    final data = await _api.me();
    return User.fromJson(data);
  }

  @override
  Future<void> signOut() async {
    try {
      final refresh = await _storage.read('auth.refresh');
      if (refresh != null) await _api.logout(refresh);
    } catch (_) {}
    await _storage.deleteAll();
  }

  @override
  Future<void> deleteAccount() async {
    await _api.deleteAccount();
    await _storage.deleteAll();
  }
}
