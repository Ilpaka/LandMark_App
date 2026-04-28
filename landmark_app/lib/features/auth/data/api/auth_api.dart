import 'package:dio/dio.dart';
import '../../../../core/networking/api_endpoints.dart';

class AuthApi {
  final Dio dio;
  AuthApi(this.dio);

  Future<Map<String, dynamic>> register(String name, String email, String password) async {
    final r = await dio.post(ApiEndpoints.authRegister, data: {'name': name, 'email': email, 'password': password});
    return r.data as Map<String, dynamic>;
  }

  Future<void> verifyEmail(String verificationId, String code) async {
    await dio.post(ApiEndpoints.authVerifyEmail, data: {'verification_id': verificationId, 'code': code});
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    final r = await dio.post(ApiEndpoints.authLogin, data: {'email': email, 'password': password});
    return r.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> me() async {
    return (await dio.get(ApiEndpoints.authMe)).data as Map<String, dynamic>;
  }

  Future<void> logout(String refreshToken) async {
    await dio.post(ApiEndpoints.authLogout, data: {'refresh_token': refreshToken});
  }

  Future<void> forgot(String email) async {
    await dio.post(ApiEndpoints.authForgotPassword, data: {'email': email});
  }

  Future<void> deleteAccount() async {
    await dio.delete(ApiEndpoints.authDeleteAccount);
  }
}
