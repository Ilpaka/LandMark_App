import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../auth/presentation/providers/auth_provider.dart';

class SplashPage extends ConsumerStatefulWidget {
  const SplashPage({super.key});

  @override
  ConsumerState<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends ConsumerState<SplashPage> {
  @override
  void initState() {
    super.initState();
    _navigate();
  }

  Future<void> _navigate() async {
    await Future.delayed(const Duration(milliseconds: 1500));
    if (!mounted) return;
    final authState = ref.read(authProvider);
    if (authState is AuthStateAuthenticated) {
      context.go('/main/map');
    } else {
      context.go('/auth/welcome');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.primary,
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.explore, size: 80, color: Colors.white),
            const SizedBox(height: AppSpacing.md),
            Text('Wanderlog',
                style: AppTypography.h1.copyWith(color: Colors.white)),
            const SizedBox(height: AppSpacing.xxl),
            const CircularProgressIndicator(color: Colors.white),
          ],
        ),
      ),
    );
  }
}
