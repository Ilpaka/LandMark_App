import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/favorites_provider.dart';
import '../../../../core/design/tokens.dart';

class FavoriteButton extends ConsumerWidget {
  final String targetType;
  final String targetId;
  final double size;

  const FavoriteButton({
    super.key,
    required this.targetType,
    required this.targetId,
    this.size = 24,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final key = (targetType, targetId);
    final state = ref.watch(isFavoriteProvider(key));

    return state.when(
      loading: () => SizedBox(
        width: size,
        height: size,
        child: const CircularProgressIndicator(
            strokeWidth: 2, color: AppColors.primary),
      ),
      error: (_, __) =>
          Icon(Icons.favorite_border, size: size, color: Colors.grey),
      data: (isFav) => GestureDetector(
        onTap: () => ref.read(isFavoriteProvider(key).notifier).toggle(),
        child: Icon(
          isFav ? Icons.favorite : Icons.favorite_border,
          size: size,
          color: isFav ? Colors.red : Colors.grey,
        ),
      ),
    );
  }
}
