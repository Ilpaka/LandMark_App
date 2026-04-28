class AppEnv {
  static String get apiBaseUrl {
    // Android emulator uses 10.0.2.2 to reach localhost
    // iOS simulator uses localhost
    return const String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'http://10.0.2.2:8080',
    );
  }
}
