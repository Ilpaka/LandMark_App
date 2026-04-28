class User {
  final String id;
  final String role;
  final bool emailVerified;
  final String? email;
  final String? phoneE164;

  const User({
    required this.id,
    required this.role,
    required this.emailVerified,
    this.email,
    this.phoneE164,
  });

  factory User.fromJson(Map<String, dynamic> j) => User(
        id: j['id'] as String,
        role: j['role'] as String? ?? 'user',
        emailVerified: j['email_verified'] as bool? ?? false,
        email: j['email'] as String?,
        phoneE164: j['phone_e164'] as String?,
      );
}
