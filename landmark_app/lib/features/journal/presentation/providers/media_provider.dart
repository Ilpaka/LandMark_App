import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/api/media_api.dart';
import '../../../../core/networking/api_client.dart';

final mediaApiProvider = Provider<MediaApi>(
  (ref) => MediaApi(ref.read(apiClientProvider).dio),
);

/// Список mediaId для конкретной записи.
final entryMediaProvider =
    FutureProvider.autoDispose.family<List<String>, String>((ref, entryId) {
  return ref.read(mediaApiProvider).getEntryMedia(entryId);
});

/// Presigned URL для конкретного mediaId.
final mediaUrlProvider =
    FutureProvider.autoDispose.family<String, String>((ref, mediaId) {
  return ref.read(mediaApiProvider).getMediaUrl(mediaId);
});
