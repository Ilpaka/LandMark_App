import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:share_plus/share_plus.dart' show Share;
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../domain/entities/entry.dart';
import '../widgets/reaction_bar.dart';

class EntryDetailPage extends ConsumerWidget {
  final JournalEntry entry;
  const EntryDetailPage({super.key, required this.entry});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final fmt = DateFormat('d MMMM yyyy', 'ru_RU');

    return Scaffold(
      appBar: AppBar(
        title: Text(entry.title ?? 'Запись'),
        actions: [
          IconButton(
            icon: const Icon(Icons.share_outlined),
            tooltip: 'Поделиться',
            onPressed: () => _share(entry),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header row
            Row(
              children: [
                const Icon(Icons.calendar_today_outlined,
                    size: 16, color: AppColors.textSecondary),
                const SizedBox(width: AppSpacing.xs),
                Text(fmt.format(entry.occurredAt), style: AppTypography.caption),
                const Spacer(),
                if (entry.mood != null)
                  Text(_moodEmoji(entry.mood!), style: const TextStyle(fontSize: 20)),
              ],
            ),
            const SizedBox(height: AppSpacing.md),

            // Title
            if (entry.title != null) ...[
              Text(entry.title!, style: AppTypography.h2),
              const SizedBox(height: AppSpacing.md),
            ],

            // Body
            Text(entry.body, style: AppTypography.body),
            const SizedBox(height: AppSpacing.xl),

            // Reactions
            const Divider(height: 1),
            const SizedBox(height: AppSpacing.md),
            const Text('Реакции',
                style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14,
                    color: AppColors.textSecondary)),
            const SizedBox(height: AppSpacing.sm),
            ReactionBar(entryId: entry.id),
            const SizedBox(height: AppSpacing.xl),
          ],
        ),
      ),
    );
  }

  void _share(JournalEntry entry) {
    final text = StringBuffer();
    if (entry.title != null) {
      text.writeln(entry.title);
      text.writeln();
    }
    text.write(entry.body);
    Share.share(text.toString());
  }

  String _moodEmoji(String mood) => switch (mood) {
    'happy' => '😊',
    'excited' => '🎉',
    'calm' => '😌',
    'tired' => '😴',
    'sad' => '😢',
    _ => '😊',
  };
}
