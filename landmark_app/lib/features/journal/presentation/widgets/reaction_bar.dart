import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../providers/reactions_provider.dart';

const _kEmojis = ['👍', '❤️', '😍', '🔥', '😂', '😮'];

class ReactionBar extends ConsumerWidget {
  final String entryId;
  const ReactionBar({super.key, required this.entryId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final reactionsAsync = ref.watch(reactionsProvider(entryId));

    return reactionsAsync.when(
      loading: () => const SizedBox(height: 40),
      error: (_, __) => const SizedBox.shrink(),
      data: (reactions) => Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Existing reaction counts
          if (reactions.counts.isNotEmpty)
            Wrap(
              spacing: AppSpacing.sm,
              runSpacing: AppSpacing.xs,
              children: reactions.counts.map((rc) {
                final isActive = reactions.mine.contains(rc.emoji);
                return _ReactionChip(
                  emoji: rc.emoji,
                  count: rc.count,
                  isActive: isActive,
                  onTap: () => ref.read(reactionsProvider(entryId).notifier).toggle(rc.emoji),
                );
              }).toList(),
            ),
          // Add reaction picker
          GestureDetector(
            onTap: () => _showPicker(context, ref, reactions.mine),
            child: Container(
              margin: const EdgeInsets.only(top: AppSpacing.xs),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                border: Border.all(color: AppColors.border),
                borderRadius: BorderRadius.circular(20),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.add, size: 14, color: AppColors.textSecondary),
                  SizedBox(width: 2),
                  Text('😊', style: TextStyle(fontSize: 14)),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _showPicker(BuildContext context, WidgetRef ref, List<String> mine) {
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Реакция', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
            const SizedBox(height: AppSpacing.md),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: _kEmojis.map((e) {
                final isActive = mine.contains(e);
                return GestureDetector(
                  onTap: () {
                    Navigator.pop(context);
                    ref.read(reactionsProvider(entryId).notifier).toggle(e);
                  },
                  child: Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: isActive ? AppColors.primary.withValues(alpha: 0.1) : null,
                      borderRadius: BorderRadius.circular(12),
                      border: isActive ? Border.all(color: AppColors.primary) : null,
                    ),
                    child: Text(e, style: const TextStyle(fontSize: 28)),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: AppSpacing.md),
          ],
        ),
      ),
    );
  }
}

class _ReactionChip extends StatelessWidget {
  final String emoji;
  final int count;
  final bool isActive;
  final VoidCallback onTap;

  const _ReactionChip({
    required this.emoji,
    required this.count,
    required this.isActive,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
        decoration: BoxDecoration(
          color: isActive ? AppColors.primary.withValues(alpha: 0.1) : AppColors.surface,
          border: Border.all(
            color: isActive ? AppColors.primary : AppColors.border,
          ),
          borderRadius: BorderRadius.circular(20),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(emoji, style: const TextStyle(fontSize: 16)),
            const SizedBox(width: 4),
            Text(
              '$count',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: isActive ? AppColors.primary : AppColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
