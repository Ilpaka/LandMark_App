import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/connectivity/connectivity_provider.dart';
import '../../../journal/data/local/draft_storage.dart';
import '../../../journal/data/sync/draft_sync_service.dart';

final _draftsProvider = Provider<List<JournalDraft>>((ref) {
  return ref.watch(draftStorageProvider).getDrafts();
});

class SyncPage extends ConsumerWidget {
  const SyncPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final drafts = ref.watch(_draftsProvider);
    final isOffline = ref.watch(isOfflineProvider);
    final fmt = DateFormat('d MMM HH:mm', 'ru_RU');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Синхронизация'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          if (!isOffline && drafts.isNotEmpty)
            TextButton(
              onPressed: () => ref.read(draftSyncProvider).flush(),
              child: const Text('Синхр.', style: TextStyle(color: AppColors.primary)),
            ),
        ],
      ),
      body: Column(
        children: [
          // Status banner
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            color: isOffline
                ? Colors.orange.shade700
                : (drafts.isEmpty ? Colors.green.shade600 : AppColors.primary),
            child: Row(
              children: [
                Icon(
                  isOffline ? Icons.wifi_off : (drafts.isEmpty ? Icons.check_circle : Icons.sync),
                  color: Colors.white,
                  size: 18,
                ),
                const SizedBox(width: 8),
                Text(
                  isOffline
                      ? 'Нет подключения — черновики ожидают'
                      : (drafts.isEmpty
                          ? 'Все записи синхронизированы'
                          : '${drafts.length} черновик(ов) ожидает отправки'),
                  style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w500),
                ),
              ],
            ),
          ),
          if (drafts.isEmpty)
            const Expanded(
              child: Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.cloud_done_outlined, size: 64, color: AppColors.textSecondary),
                    SizedBox(height: 12),
                    Text(
                      'Нет неотправленных черновиков',
                      style: TextStyle(fontSize: 16, color: AppColors.textSecondary),
                    ),
                  ],
                ),
              ),
            )
          else
            Expanded(
              child: ListView.separated(
                padding: const EdgeInsets.all(16),
                itemCount: drafts.length,
                separatorBuilder: (_, __) => const SizedBox(height: 8),
                itemBuilder: (_, i) {
                  final draft = drafts[i];
                  return Card(
                    elevation: 0,
                    color: AppColors.surface,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    child: ListTile(
                      leading: Container(
                        width: 40,
                        height: 40,
                        decoration: BoxDecoration(
                          color: AppColors.primary.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: const Icon(Icons.article_outlined, color: AppColors.primary, size: 22),
                      ),
                      title: Text(
                        draft.title.isNotEmpty ? draft.title : 'Без названия',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontWeight: FontWeight.w500),
                      ),
                      subtitle: Text(
                        fmt.format(draft.createdAt),
                        style: const TextStyle(fontSize: 12, color: AppColors.textSecondary),
                      ),
                      trailing: IconButton(
                        icon: const Icon(Icons.delete_outline, color: Colors.red, size: 20),
                        onPressed: () async {
                          await ref.read(draftStorageProvider).removeDraft(draft.id);
                          ref.invalidate(_draftsProvider);
                        },
                      ),
                    ),
                  );
                },
              ),
            ),
        ],
      ),
    );
  }
}
