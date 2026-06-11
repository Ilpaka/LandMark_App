import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';

final reverseGeocodingProvider =
    Provider<ReverseGeocodingService>((ref) => ReverseGeocodingService());

class ReverseGeocodeResult {
  final String? city;
  final String? country;
  const ReverseGeocodeResult({this.city, this.country});
}

/// Обратное геокодирование через OSM Nominatim (без ключа).
///
/// Политика Nominatim: не чаще 1 запроса в секунду и обязательный
/// осмысленный User-Agent. Мы запрашиваем один раз по факту выбора точки
/// в пикере, что укладывается в лимиты.
class ReverseGeocodingService {
  ReverseGeocodingService([Dio? dio])
      : _dio = dio ??
            Dio(BaseOptions(
              baseUrl: 'https://nominatim.openstreetmap.org',
              connectTimeout: const Duration(seconds: 8),
              receiveTimeout: const Duration(seconds: 8),
              headers: {'User-Agent': 'com.landmark.app (LandMark)'},
            ));

  final Dio _dio;

  /// Определяет город и страну для точки. При любой ошибке сети возвращает
  /// пустой результат — автозаполнение не должно ломать форму.
  Future<ReverseGeocodeResult> resolve(LatLng point) async {
    try {
      final r = await _dio.get<Map<String, dynamic>>(
        '/reverse',
        queryParameters: {
          'format': 'jsonv2',
          'lat': point.latitude,
          'lon': point.longitude,
          'zoom': 10,
          'addressdetails': 1,
          'accept-language': 'ru',
        },
      );
      final addr = (r.data?['address'] as Map<String, dynamic>?) ?? const {};
      final city = (addr['city'] ??
              addr['town'] ??
              addr['village'] ??
              addr['municipality'] ??
              addr['county'] ??
              addr['state'])
          ?.toString();
      return ReverseGeocodeResult(
        city: city,
        country: addr['country']?.toString(),
      );
    } on DioException {
      return const ReverseGeocodeResult();
    }
  }
}
