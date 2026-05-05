import 'package:dio/dio.dart';

class AdminUser {
  final String id;
  final String? email;
  final String? phoneE164;
  final String status; // active | pending_verification | blocked | deleted | new
  final String role;   // user | admin | system
  final DateTime? emailVerifiedAt;
  final DateTime? phoneVerifiedAt;
  final DateTime? blockedAt;
  final String? blockedReason;
  final DateTime? lastLoginAt;
  final DateTime createdAt;

  const AdminUser({
    required this.id,
    required this.status,
    required this.role,
    required this.createdAt,
    this.email,
    this.phoneE164,
    this.emailVerifiedAt,
    this.phoneVerifiedAt,
    this.blockedAt,
    this.blockedReason,
    this.lastLoginAt,
  });

  factory AdminUser.fromJson(Map<String, dynamic> j) => AdminUser(
        id: j['id'] as String,
        email: j['email'] as String?,
        phoneE164: j['phone_e164'] as String?,
        status: j['status'] as String? ?? 'active',
        role: j['role'] as String? ?? 'user',
        emailVerifiedAt: _t(j['email_verified_at']),
        phoneVerifiedAt: _t(j['phone_verified_at']),
        blockedAt: _t(j['blocked_at']),
        blockedReason: j['blocked_reason'] as String?,
        lastLoginAt: _t(j['last_login_at']),
        createdAt: DateTime.parse(j['created_at'] as String),
      );

  bool get isBlocked => status == 'blocked';
  bool get isAdmin => role == 'admin';
  String get displayLabel =>
      email ?? phoneE164 ?? id.substring(0, 8);
}

class AdminAuditEvent {
  final String id;
  final String? userId;
  final String eventType;
  final String? ip;
  final Map<String, dynamic>? meta;
  final DateTime createdAt;

  const AdminAuditEvent({
    required this.id,
    required this.eventType,
    required this.createdAt,
    this.userId,
    this.ip,
    this.meta,
  });

  factory AdminAuditEvent.fromJson(Map<String, dynamic> j) => AdminAuditEvent(
        id: j['id'] as String,
        userId: j['user_id'] as String?,
        eventType: j['event_type'] as String,
        ip: j['ip'] as String?,
        meta: (j['meta'] as Map?)?.cast<String, dynamic>(),
        createdAt: DateTime.parse(j['created_at'] as String),
      );
}

DateTime? _t(dynamic v) {
  if (v == null) return null;
  if (v is String && v.isEmpty) return null;
  return DateTime.tryParse(v as String);
}

class UsersApi {
  final Dio _dio;
  UsersApi(this._dio);

  Future<List<AdminUser>> list({String? query, int limit = 50, int offset = 0}) async {
    final resp = await _dio.get('/v1/admin/users', queryParameters: {
      if (query != null && query.isNotEmpty) 'q': query,
      'limit': limit,
      'offset': offset,
    });
    final data = resp.data as Map<String, dynamic>;
    final list = (data['users'] as List).cast<Map<String, dynamic>>();
    return list.map(AdminUser.fromJson).toList();
  }

  Future<void> block(String userId, String reason) async {
    await _dio.post('/v1/admin/users/$userId/block', data: {'reason': reason});
  }

  Future<void> unblock(String userId) async {
    await _dio.post('/v1/admin/users/$userId/unblock');
  }

  Future<void> forceLogout(String userId) async {
    await _dio.post('/v1/admin/users/$userId/force-logout');
  }

  Future<List<AdminAuditEvent>> audit(String userId) async {
    final resp = await _dio.get('/v1/admin/users/$userId/audit');
    final data = resp.data as Map<String, dynamic>;
    final list = (data['events'] as List).cast<Map<String, dynamic>>();
    return list.map(AdminAuditEvent.fromJson).toList();
  }
}
