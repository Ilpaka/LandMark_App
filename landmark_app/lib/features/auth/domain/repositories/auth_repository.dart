import '../entities/user.dart';

abstract class AuthRepository {
  Future<User?> tryRestoreSession();
  Future<String> register({required String name, required String email, required String password});
  Future<void> verifyEmail({required String verificationId, required String code});
  Future<User> login({required String email, required String password});
  Future<void> forgotPassword({required String email});
  Future<User> me();
  Future<void> signOut();
  Future<void> deleteAccount();
}
