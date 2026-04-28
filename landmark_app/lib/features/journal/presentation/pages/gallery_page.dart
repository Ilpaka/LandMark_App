import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../providers/media_provider.dart';
import 'photo_view_page.dart';

class GalleryPage extends ConsumerWidget {
  final String entryId;
  final String? entryTitle;

  const GalleryPage({super.key, required this.entryId, this.entryTitle});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final mediaAsync = ref.watch(entryMediaProvider(entryId));

    return Scaffold(
      appBar: AppBar(
        title: Text(entryTitle ?? 'Галерея'),
      ),
      body: mediaAsync.when(
        loading: () => const Center(
          child: CircularProgressIndicator(color: AppColors.primary),
        ),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (mediaIds) => mediaIds.isEmpty
            ? const _EmptyGallery()
            : GridView.builder(
                padding: const EdgeInsets.all(AppSpacing.xs),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 3,
                  mainAxisSpacing: 2,
                  crossAxisSpacing: 2,
                ),
                itemCount: mediaIds.length,
                itemBuilder: (_, i) => _Thumbnail(
                  mediaId: mediaIds[i],
                  onTap: () => Navigator.push<void>(
                    context,
                    MaterialPageRoute(
                      builder: (_) => PhotoViewPage(
                        mediaIds: mediaIds,
                        initialIndex: i,
                      ),
                    ),
                  ),
                ),
              ),
      ),
    );
  }
}

class _Thumbnail extends ConsumerWidget {
  final String mediaId;
  final VoidCallback onTap;

  const _Thumbnail({required this.mediaId, required this.onTap});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final urlAsync = ref.watch(mediaUrlProvider(mediaId));

    return GestureDetector(
      onTap: onTap,
      child: urlAsync.when(
        loading: () => Container(color: AppColors.surface),
        error: (_, __) => Container(
          color: AppColors.surface,
          child: const Icon(Icons.broken_image_outlined,
              color: AppColors.textSecondary),
        ),
        data: (url) => CachedNetworkImage(
          imageUrl: url,
          fit: BoxFit.cover,
          placeholder: (_, __) => Container(color: AppColors.surface),
          errorWidget: (_, __, ___) => Container(
            color: AppColors.surface,
            child: const Icon(Icons.broken_image_outlined,
                color: AppColors.textSecondary),
          ),
        ),
      ),
    );
  }
}

class _EmptyGallery extends StatelessWidget {
  const _EmptyGallery();

  @override
  Widget build(BuildContext context) => const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.photo_library_outlined,
                size: 64, color: AppColors.textSecondary),
            SizedBox(height: 16),
            Text('Нет фотографий',
                style: TextStyle(
                    fontSize: 18, color: AppColors.textSecondary)),
            SizedBox(height: 8),
            Text('Добавьте фото при создании записи',
                style: TextStyle(color: AppColors.textSecondary)),
          ],
        ),
      );
}
