import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../features/auth/presentation/pages/welcome_page.dart';
import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/auth/presentation/pages/register_page.dart';
import '../../features/auth/presentation/pages/otp_page.dart';
import '../../features/auth/presentation/providers/auth_provider.dart';
import '../../features/splash/presentation/pages/splash_page.dart';
import '../../features/main_shell/presentation/pages/main_shell_page.dart';
import '../../features/map/presentation/pages/map_page.dart';
import '../../features/trips/presentation/pages/trips_list_page.dart';
import '../../features/journal/presentation/pages/journal_feed_page.dart';
import '../../features/admin/presentation/pages/moderation_queue_page.dart';
import '../../features/profile/presentation/pages/profile_page.dart';

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
          GoRoute(path: '/main/map', builder: (_, __) => const MapPage()),
          GoRoute(path: '/main/trips', builder: (_, __) => const TripsListPage()),
          GoRoute(path: '/main/journal', builder: (_, __) => const JournalFeedPage()),
          GoRoute(path: '/main/profile', builder: (_, __) => const ProfilePage()),
          GoRoute(path: '/main/admin', builder: (_, __) => const ModerationQueuePage()),
        ],
      ),
    ],
  );
});

