import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../journal/presentation/providers/media_provider.dart';

class UserAvatar extends ConsumerWidget {
  final String? mediaId;
  final String displayName;
  final double radius;

  const UserAvatar({
    super.key,
    required this.displayName,
    this.mediaId,
    this.radius = 24,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final initial =
        (displayName.isNotEmpty ? displayName[0] : '?').toUpperCase();
    final fontSize = radius * 0.7;

    if (mediaId == null) {
      return _Fallback(initial: initial, radius: radius, fontSize: fontSize);
    }

    final urlAsync = ref.watch(mediaUrlProvider(mediaId!));
    return urlAsync.when(
      loading: () =>
          _Fallback(initial: initial, radius: radius, fontSize: fontSize),
      error: (_, __) =>
          _Fallback(initial: initial, radius: radius, fontSize: fontSize),
      data: (url) => CircleAvatar(
        radius: radius,
        backgroundColor: AppColors.primary.withValues(alpha: 0.1),
        child: ClipOval(
          child: CachedNetworkImage(
            imageUrl: url,
            width: radius * 2,
            height: radius * 2,
            fit: BoxFit.cover,
            placeholder: (_, __) =>
                _Fallback(initial: initial, radius: radius, fontSize: fontSize),
            errorWidget: (_, __, ___) =>
                _Fallback(initial: initial, radius: radius, fontSize: fontSize),
          ),
        ),
      ),
    );
  }
}

class _Fallback extends StatelessWidget {
  final String initial;
  final double radius;
  final double fontSize;
  const _Fallback(
      {required this.initial, required this.radius, required this.fontSize});

  @override
  Widget build(BuildContext context) => CircleAvatar(
        radius: radius,
        backgroundColor: AppColors.primary.withValues(alpha: 0.15),
        child: Text(
          initial,
          style: TextStyle(
            fontSize: fontSize,
            color: AppColors.primary,
            fontWeight: FontWeight.w600,
          ),
        ),
      );
}
