import 'package:flutter_local_notifications/flutter_local_notifications.dart';

/// Демо-помощник: при регистрации показывает локальный пуш с OTP-кодом,
/// чтобы код подтверждения видно было прямо на устройстве, без почтового сервера.
class OtpNotifier {
  static final FlutterLocalNotificationsPlugin _plugin =
      FlutterLocalNotificationsPlugin();
  static bool _initialized = false;

  static Future<void> _ensureInitialized() async {
    if (_initialized) return;
    const androidInit = AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosInit = DarwinInitializationSettings(
      requestAlertPermission: true,
      requestBadgePermission: false,
      requestSoundPermission: false,
    );
    await _plugin.initialize(
      const InitializationSettings(android: androidInit, iOS: iosInit),
    );
    await _plugin
        .resolvePlatformSpecificImplementation<
            IOSFlutterLocalNotificationsPlugin>()
        ?.requestPermissions(alert: true, badge: false, sound: true);
    await _plugin
        .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin>()
        ?.createNotificationChannel(const AndroidNotificationChannel(
          'otp',
          'Коды подтверждения',
          description: 'Демо-уведомления с OTP-кодами',
          importance: Importance.high,
        ));
    _initialized = true;
  }

  static Future<void> showOtp(String code) async {
    await _ensureInitialized();
    await _plugin.show(
      1001,
      'Код подтверждения LandMark',
      'Ваш код: $code',
      const NotificationDetails(
        android: AndroidNotificationDetails(
          'otp',
          'Коды подтверждения',
          importance: Importance.high,
          priority: Priority.high,
        ),
        iOS: DarwinNotificationDetails(
          presentAlert: true,
          presentBanner: true,
          presentList: true,
          presentSound: true,
        ),
      ),
    );
  }
}
