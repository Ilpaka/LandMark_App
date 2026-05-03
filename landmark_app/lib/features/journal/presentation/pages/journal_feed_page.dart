import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../domain/entities/entry.dart';
import '../providers/journal_provider.dart';
import '../../../../core/widgets/skeleton.dart';
import 'create_entry_page.dart';
import 'entry_detail_page.dart';

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
            : _PaginatedList(entries: entries),
        loading: () => const SkeletonListView(),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
      ),
    );
  }
}

class _PaginatedList extends ConsumerWidget {
  final List<JournalEntry> entries;
  const _PaginatedList({required this.entries});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifier = ref.read(journalFeedProvider.notifier);

    return NotificationListener<ScrollNotification>(
      onNotification: (n) {
        if (n is ScrollEndNotification &&
            n.metrics.extentAfter < 200 &&
            notifier.hasMore) {
          notifier.loadMore();
        }
        return false;
      },
      child: ListView.builder(
        padding: const EdgeInsets.all(AppSpacing.md),
        itemCount: entries.length + 1,
        itemBuilder: (_, i) {
          if (i == entries.length) {
            return notifier.hasMore
                ? const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Center(
                        child: CircularProgressIndicator(
                            color: AppColors.primary, strokeWidth: 2)))
                : const SizedBox.shrink();
          }
          return _EntryCard(entry: entries[i]);
        },
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
    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.md),
      child: InkWell(
        onTap: () => Navigator.push(
          context,
          MaterialPageRoute<void>(
              builder: (_) => EntryDetailPage(entry: entry)),
        ),
        borderRadius: const BorderRadius.all(AppRadius.md),
        child: Card(
          margin: EdgeInsets.zero,
          shape: const RoundedRectangleBorder(
              borderRadius: BorderRadius.all(AppRadius.md)),
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.md),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        entry.title ??
                            'Запись от ${fmt.format(entry.occurredAt)}',
                        style: AppTypography.h3,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (entry.mood != null)
                      Text(_moodEmoji(entry.mood!),
                          style: const TextStyle(fontSize: 20)),
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
                Text(fmt.format(entry.occurredAt),
                    style: AppTypography.caption),
              ],
            ),
          ),
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
