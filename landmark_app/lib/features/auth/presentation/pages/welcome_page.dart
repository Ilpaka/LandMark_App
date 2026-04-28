import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';

class WelcomePage extends StatelessWidget {
  const WelcomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.primary,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Spacer(),
              const Icon(Icons.explore, size: 64, color: Colors.white),
              const SizedBox(height: AppSpacing.md),
              Text(
                'Wanderlog',
                style: AppTypography.h1.copyWith(color: Colors.white),
              ),
              const SizedBox(height: AppSpacing.sm),
              Text(
                'Журнал путешественника\nс картой достопримечательностей',
                style: AppTypography.body.copyWith(color: Colors.white70),
              ),
              const Spacer(),
              ElevatedButton(
                onPressed: () => context.go('/auth/register'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.white,
                  foregroundColor: AppColors.primary,
                  minimumSize: const Size(double.infinity, 52),
                  shape: const RoundedRectangleBorder(
                    borderRadius: BorderRadius.all(AppRadius.md),
                  ),
                ),
                child: const Text('Начать', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
              ),
              const SizedBox(height: AppSpacing.md),
              TextButton(
                onPressed: () => context.go('/auth/login'),
                child: const Text(
                  'Уже есть аккаунт',
                  style: TextStyle(color: Colors.white),
                ),
              ),
              const SizedBox(height: AppSpacing.md),
            ],
          ),
        ),
      ),
    );
  }
}
