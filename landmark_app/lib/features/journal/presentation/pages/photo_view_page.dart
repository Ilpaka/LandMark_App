import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:share_plus/share_plus.dart';
import '../../../../core/design/tokens.dart';
import '../providers/media_provider.dart';

class PhotoViewPage extends ConsumerStatefulWidget {
  final List<String> mediaIds;
  final int initialIndex;

  const PhotoViewPage({
    super.key,
    required this.mediaIds,
    this.initialIndex = 0,
  });

  @override
  ConsumerState<PhotoViewPage> createState() => _PhotoViewPageState();
}

class _PhotoViewPageState extends ConsumerState<PhotoViewPage> {
  late final PageController _pageCtrl;
  late int _current;

  @override
  void initState() {
    super.initState();
    _current = widget.initialIndex;
    _pageCtrl = PageController(initialPage: widget.initialIndex);
  }

  @override
  void dispose() {
    _pageCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      extendBodyBehindAppBar: true,
      appBar: AppBar(
        backgroundColor: Colors.black54,
        foregroundColor: Colors.white,
        elevation: 0,
        title: Text(
          '${_current + 1} / ${widget.mediaIds.length}',
          style: const TextStyle(color: Colors.white),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.share_outlined, color: Colors.white),
            tooltip: 'Поделиться',
            onPressed: () => _shareCurrentPhoto(),
          ),
        ],
      ),
      body: PageView.builder(
        controller: _pageCtrl,
        itemCount: widget.mediaIds.length,
        onPageChanged: (i) => setState(() => _current = i),
        itemBuilder: (_, i) => _PhotoSlide(mediaId: widget.mediaIds[i]),
      ),
    );
  }

  void _shareCurrentPhoto() {
    final urlAsync = ref.read(mediaUrlProvider(widget.mediaIds[_current]));
    urlAsync.whenData((url) => Share.share(url));
  }
}

class _PhotoSlide extends ConsumerWidget {
  final String mediaId;
  const _PhotoSlide({required this.mediaId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final urlAsync = ref.watch(mediaUrlProvider(mediaId));

    return urlAsync.when(
      loading: () => const Center(
        child: CircularProgressIndicator(color: AppColors.primary),
      ),
      error: (_, __) => const Center(
        child:
            Icon(Icons.broken_image_outlined, color: Colors.white54, size: 64),
      ),
      data: (url) => InteractiveViewer(
        minScale: 0.8,
        maxScale: 5.0,
        child: Center(
          child: CachedNetworkImage(
            imageUrl: url,
            fit: BoxFit.contain,
            placeholder: (_, __) => const Center(
              child: CircularProgressIndicator(color: AppColors.primary),
            ),
            errorWidget: (_, __, ___) => const Icon(
              Icons.broken_image_outlined,
              color: Colors.white54,
              size: 64,
            ),
          ),
        ),
      ),
    );
  }
}
