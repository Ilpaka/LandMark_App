import 'dart:io' show Platform;
import 'package:flutter/foundation.dart' show kIsWeb;

class AppEnv {
  static String get apiBaseUrl {
    const fromEnv = String.fromEnvironment('API_BASE_URL');
    if (fromEnv.isNotEmpty) return fromEnv;
    // Android эмулятор → 10.0.2.2, iOS симулятор и десктоп → localhost
    if (!kIsWeb && Platform.isAndroid) return 'http://10.0.2.2:8080';
    return 'http://localhost:8080';
  }
}
