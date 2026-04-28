import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../domain/entities/entry.dart';
import '../providers/journal_provider.dart';
import 'create_entry_page.dart';

class JournalFeedPage extends ConsumerWidget {
  const JournalFeedPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final entriesAsync = ref.watch(journalFeedProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Журнал')),
      floatingActionButton: FloatingActionButton(
        backgroundColor: AppColors.primary,
        onPressed: () async {
          final result = await Navigator.push<bool>(
            context,
            MaterialPageRoute(builder: (_) => const CreateEntryPage()),
          );
          if (result == true) ref.invalidate(journalFeedProvider);
        },
        child: const Icon(Icons.add, color: Colors.white),
      ),
      body: entriesAsync.when(
        data: (entries) => entries.isEmpty
            ? const _EmptyJournal()
            : ListView.builder(
                padding: const EdgeInsets.all(AppSpacing.md),
                itemCount: entries.length,
                itemBuilder: (_, i) => _EntryCard(entry: entries[i]),
              ),
        loading: () => const Center(child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
      ),
    );
  }
}

class _EntryCard extends StatelessWidget {
  final JournalEntry entry;
  const _EntryCard({required this.entry});

  @override
  Widget build(BuildContext context) {
    final fmt = DateFormat('d MMM yyyy', 'ru_RU');
    return Card(
      margin: const EdgeInsets.only(bottom: AppSpacing.md),
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.all(AppRadius.md)),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    entry.title ?? 'Запись от ${fmt.format(entry.occurredAt)}',
                    style: AppTypography.h3,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (entry.mood != null)
                  Text(_moodEmoji(entry.mood!), style: const TextStyle(fontSize: 20)),
              ],
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              entry.body,
              style: AppTypography.body,
              maxLines: 3,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(fmt.format(entry.occurredAt), style: AppTypography.caption),
          ],
        ),
      ),
    );
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

class _EmptyJournal extends StatelessWidget {
  const _EmptyJournal();

  @override
  Widget build(BuildContext context) => const Center(
    child: Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Icon(Icons.book_outlined, size: 64, color: AppColors.textSecondary),
        SizedBox(height: 16),
        Text(
          'Нет записей',
          style: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.w600,
            color: AppColors.textPrimary,
          ),
        ),
        SizedBox(height: 8),
        Text(
          'Нажмите + чтобы добавить запись',
          style: TextStyle(color: AppColors.textSecondary),
        ),
      ],
    ),
  );
}
