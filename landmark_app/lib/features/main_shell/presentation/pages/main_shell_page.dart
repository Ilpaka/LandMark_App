import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/design/tokens.dart';
import '../../../auth/presentation/providers/auth_provider.dart';

class MainShellPage extends ConsumerWidget {
  final Widget child;
  const MainShellPage({super.key, required this.child});

  bool _isAdmin(WidgetRef ref) {
    final auth = ref.watch(authProvider);
    if (auth is AuthStateAuthenticated) return auth.user.role == 'admin';
    return false;
  }

  int _selectedIndex(BuildContext context, bool isAdmin) {
    final loc = GoRouterState.of(context).matchedLocation;
    if (loc.startsWith('/main/trips')) return 1;
    if (loc.startsWith('/main/journal')) return 2;
    if (loc.startsWith('/main/favorites')) return 3;
    if (loc.startsWith('/main/profile')) return 4;
    if (isAdmin && loc.startsWith('/main/admin')) return 5;
    return 0;
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isAdmin = _isAdmin(ref);
    final idx = _selectedIndex(context, isAdmin);

    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: idx,
        backgroundColor: AppColors.surface,
        indicatorColor: AppColors.primary.withValues(alpha: 0.15),
        onDestinationSelected: (i) {
          switch (i) {
            case 0: context.go('/main/map'); break;
            case 1: context.go('/main/trips'); break;
            case 2: context.go('/main/journal'); break;
            case 3: context.go('/main/favorites'); break;
            case 4: context.go('/main/profile'); break;
            case 5: if (isAdmin) context.go('/main/admin'); break;
          }
        },
        destinations: [
          const NavigationDestination(icon: Icon(Icons.map_outlined), selectedIcon: Icon(Icons.map), label: 'Карта'),
          const NavigationDestination(icon: Icon(Icons.luggage_outlined), selectedIcon: Icon(Icons.luggage), label: 'Поездки'),
          const NavigationDestination(icon: Icon(Icons.book_outlined), selectedIcon: Icon(Icons.book), label: 'Журнал'),
          const NavigationDestination(icon: Icon(Icons.favorite_outline), selectedIcon: Icon(Icons.favorite), label: 'Избранное'),
          const NavigationDestination(icon: Icon(Icons.person_outline), selectedIcon: Icon(Icons.person), label: 'Профиль'),
          if (isAdmin)
            const NavigationDestination(
              icon: Icon(Icons.admin_panel_settings_outlined),
              selectedIcon: Icon(Icons.admin_panel_settings),
              label: 'Модерация',
            ),
        ],
      ),
    );
  }
}
