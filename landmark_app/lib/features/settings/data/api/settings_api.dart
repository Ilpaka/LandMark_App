import 'package:dio/dio.dart';

class NotificationPreferences {
  final bool pushTripReminders;
  final bool pushModerationResult;
  final bool pushMarketing;
  final bool emailModerationResult;
  final bool emailMarketing;

  const NotificationPreferences({
    required this.pushTripReminders,
    required this.pushModerationResult,
    required this.pushMarketing,
    required this.emailModerationResult,
    required this.emailMarketing,
  });

  factory NotificationPreferences.fromJson(Map<String, dynamic> j) =>
      NotificationPreferences(
        pushTripReminders: j['push_trip_reminders'] as bool? ?? true,
        pushModerationResult: j['push_moderation_result'] as bool? ?? true,
        pushMarketing: j['push_marketing'] as bool? ?? false,
        emailModerationResult: j['email_moderation_result'] as bool? ?? true,
        emailMarketing: j['email_marketing'] as bool? ?? false,
      );

  NotificationPreferences copyWith({
    bool? pushTripReminders,
    bool? pushModerationResult,
    bool? pushMarketing,
    bool? emailModerationResult,
    bool? emailMarketing,
  }) =>
      NotificationPreferences(
        pushTripReminders: pushTripReminders ?? this.pushTripReminders,
        pushModerationResult: pushModerationResult ?? this.pushModerationResult,
        pushMarketing: pushMarketing ?? this.pushMarketing,
        emailModerationResult:
            emailModerationResult ?? this.emailModerationResult,
        emailMarketing: emailMarketing ?? this.emailMarketing,
      );
}

class SettingsApi {
  final Dio dio;
  SettingsApi(this.dio);

  Future<NotificationPreferences> getNotifPrefs() async {
    final r = await dio.get('/v1/notifications/preferences');
    return NotificationPreferences.fromJson(r.data as Map<String, dynamic>);
  }

  Future<NotificationPreferences> updateNotifPrefs(
      Map<String, dynamic> patch) async {
    final r = await dio.patch('/v1/notifications/preferences', data: patch);
    return NotificationPreferences.fromJson(r.data as Map<String, dynamic>);
  }
}
