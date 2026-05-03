import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final apiClientProvider =
    Provider<ApiClient>((ref) => throw UnimplementedError());

class ApiClient {
  final Dio _dio;
  final Future<String?> Function() getRefreshToken;
  final Future<void> Function(String access, String refresh) saveTokens;
  final Future<void> Function() clearTokens;

  ApiClient({
    required String baseUrl,
    required Future<String?> Function() getAccessToken,
    required this.getRefreshToken,
    required this.saveTokens,
    required this.clearTokens,
  }) : _dio = Dio(BaseOptions(
          baseUrl: baseUrl,
          connectTimeout: const Duration(seconds: 15),
          receiveTimeout: const Duration(seconds: 30),
          headers: {'Content-Type': 'application/json'},
        )) {
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await getAccessToken();
          if (token != null && token.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
        onError: (err, handler) async {
          if (err.response?.statusCode == 401) {
            try {
              final refresh = await getRefreshToken();
              if (refresh != null && refresh.isNotEmpty) {
                final r = await _dio.post<Map<String, dynamic>>(
                    '/v1/auth/refresh',
                    data: {'refresh_token': refresh});
                final body = r.data!;
                final newAccess = body['access_token'] as String;
                final newRefresh = body['refresh_token'] as String;
                await saveTokens(newAccess, newRefresh);
                final opts = err.requestOptions;
                opts.headers['Authorization'] = 'Bearer $newAccess';
                final retry = await _dio.fetch(opts);
                handler.resolve(retry);
                return;
              }
            } catch (_) {
              await clearTokens();
            }
          }
          handler.next(err);
        },
      ),
    );
  }

  Dio get dio => _dio;
}
