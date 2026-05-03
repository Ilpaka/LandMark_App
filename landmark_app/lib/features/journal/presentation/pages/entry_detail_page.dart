import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:share_plus/share_plus.dart' show Share;
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/networking/api_client.dart';
import '../../domain/entities/entry.dart';
import '../providers/journal_provider.dart';
import '../providers/media_provider.dart';
import '../widgets/reaction_bar.dart';
import 'gallery_page.dart';
import 'photo_view_page.dart';

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
          PopupMenuButton<String>(
            onSelected: (v) => _onMenu(context, ref, v),
            itemBuilder: (_) => [
              const PopupMenuItem(
                  value: 'delete',
                  child: Text('Удалить', style: TextStyle(color: Colors.red))),
            ],
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
                Text(fmt.format(entry.occurredAt),
                    style: AppTypography.caption),
                const Spacer(),
                if (entry.mood != null)
                  Text(_moodEmoji(entry.mood!),
                      style: const TextStyle(fontSize: 20)),
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

            // Медиа-полоса
            _MediaStrip(entry: entry),

            // Reactions
            const Divider(height: 1),
            const SizedBox(height: AppSpacing.md),
            const Text('Реакции',
                style: TextStyle(
                    fontWeight: FontWeight.w600,
                    fontSize: 14,
                    color: AppColors.textSecondary)),
            const SizedBox(height: AppSpacing.sm),
            ReactionBar(entryId: entry.id),
            const SizedBox(height: AppSpacing.xl),
          ],
        ),
      ),
    );
  }

  Future<void> _onMenu(
      BuildContext context, WidgetRef ref, String action) async {
    if (action != 'delete') return;
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить запись?'),
        content: const Text('Запись будет безвозвратно удалена.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Отмена')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red, foregroundColor: Colors.white),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
    if (confirm != true || !context.mounted) return;
    try {
      await ref
          .read(apiClientProvider)
          .dio
          .delete('/v1/journal/entries/${entry.id}');
      ref.invalidate(journalFeedProvider);
      if (context.mounted) Navigator.pop(context);
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Ошибка: $e')));
      }
    }
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

// ---------------------------------------------------------------------------
// Горизонтальная полоса миниатюр (JR-04 entry point)
// ---------------------------------------------------------------------------

class _MediaStrip extends ConsumerWidget {
  final JournalEntry entry;
  const _MediaStrip({required this.entry});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final mediaAsync = ref.watch(entryMediaProvider(entry.id));

    return mediaAsync.when(
      loading: () => const SizedBox(
          height: 90,
          child: Center(
              child: CircularProgressIndicator(
                  color: AppColors.primary, strokeWidth: 2))),
      error: (_, __) => const SizedBox.shrink(),
      data: (ids) {
        if (ids.isEmpty) return const SizedBox.shrink();

        const thumbSize = 80.0;
        final preview = ids.take(5).toList();

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: AppSpacing.md),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Фото (${ids.length})',
                    style: const TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: 14,
                        color: AppColors.textSecondary)),
                TextButton(
                  onPressed: () => Navigator.push<void>(
                    context,
                    MaterialPageRoute(
                        builder: (_) => GalleryPage(
                              entryId: entry.id,
                              entryTitle: entry.title,
                            )),
                  ),
                  child: const Text('Все фото',
                      style: TextStyle(color: AppColors.primary)),
                ),
              ],
            ),
            SizedBox(
              height: thumbSize,
              child: ListView.separated(
                scrollDirection: Axis.horizontal,
                padding: EdgeInsets.zero,
                itemCount: preview.length,
                separatorBuilder: (_, __) =>
                    const SizedBox(width: AppSpacing.xs),
                itemBuilder: (_, i) {
                  final isLast = i == 4 && ids.length > 5;
                  return GestureDetector(
                    onTap: () => Navigator.push<void>(
                      context,
                      MaterialPageRoute(
                          builder: (_) => PhotoViewPage(
                                mediaIds: ids,
                                initialIndex: i,
                              )),
                    ),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Stack(
                        children: [
                          _ThumbImage(mediaId: preview[i], size: thumbSize),
                          if (isLast)
                            Container(
                              width: thumbSize,
                              height: thumbSize,
                              color: Colors.black54,
                              alignment: Alignment.center,
                              child: Text('+${ids.length - 4}',
                                  style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 18,
                                      fontWeight: FontWeight.w700)),
                            ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
            const SizedBox(height: AppSpacing.md),
            const Divider(height: 1),
          ],
        );
      },
    );
  }
}

class _ThumbImage extends ConsumerWidget {
  final String mediaId;
  final double size;
  const _ThumbImage({required this.mediaId, required this.size});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final urlAsync = ref.watch(mediaUrlProvider(mediaId));
    return SizedBox(
      width: size,
      height: size,
      child: urlAsync.when(
        loading: () => Container(color: AppColors.surface),
        error: (_, __) => Container(
          color: AppColors.surface,
          child: const Icon(Icons.broken_image_outlined,
              color: AppColors.textSecondary, size: 28),
        ),
        data: (url) => CachedNetworkImage(
          imageUrl: url,
          fit: BoxFit.cover,
          placeholder: (_, __) => Container(color: AppColors.surface),
          errorWidget: (_, __, ___) => Container(
            color: AppColors.surface,
            child: const Icon(Icons.broken_image_outlined,
                color: AppColors.textSecondary, size: 28),
          ),
        ),
      ),
    );
  }
}
