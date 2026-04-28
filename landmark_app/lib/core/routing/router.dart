import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../features/auth/presentation/pages/welcome_page.dart';
import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/auth/presentation/pages/register_page.dart';
import '../../features/auth/presentation/pages/otp_page.dart';
import '../../features/auth/presentation/providers/auth_provider.dart';
import '../../features/splash/presentation/pages/splash_page.dart';
import '../../features/main_shell/presentation/pages/main_shell_page.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final refreshListenable = ref.read(authRouterNotifierProvider);
  return GoRouter(
    initialLocation: '/',
    redirect: (context, state) {
      final authState = ref.read(authProvider);
      final loc = state.matchedLocation;
      final isUnknown = authState is AuthStateUnknown;
      final isAuthed = authState is AuthStateAuthenticated;
      final inAuthFlow = loc.startsWith('/auth') || loc == '/';

      if (isUnknown) return null;
      if (!isAuthed && !inAuthFlow) return '/auth/welcome';
      if (isAuthed && inAuthFlow) return '/main/map';
      return null;
    },
    refreshListenable: refreshListenable,
    routes: [
      GoRoute(path: '/', builder: (_, __) => const SplashPage()),
      GoRoute(path: '/auth/welcome', builder: (_, __) => const WelcomePage()),
      GoRoute(path: '/auth/login', builder: (_, __) => const LoginPage()),
      GoRoute(path: '/auth/register', builder: (_, __) => const RegisterPage()),
      GoRoute(
        path: '/auth/otp',
        builder: (_, state) {
          final extra = state.extra as Map<String, String>?;
          return OtpPage(
            verificationId: extra?['verification_id'] ?? '',
            kind: extra?['kind'] ?? 'registration',
          );
        },
      ),
      ShellRoute(
        builder: (_, __, child) => MainShellPage(child: child),
        routes: [
          GoRoute(path: '/main/map', builder: (_, __) => const _StubPage(label: 'Карта')),
          GoRoute(path: '/main/trips', builder: (_, __) => const _StubPage(label: 'Поездки')),
          GoRoute(path: '/main/journal', builder: (_, __) => const _StubPage(label: 'Журнал')),
          GoRoute(path: '/main/profile', builder: (_, __) => const _StubPage(label: 'Профиль')),
        ],
      ),
    ],
  );
});

class _StubPage extends StatelessWidget {
  final String label;
  const _StubPage({required this.label});
  @override
  Widget build(BuildContext context) {
    return Scaffold(body: Center(child: Text(label, style: const TextStyle(fontSize: 24))));
  }
}
