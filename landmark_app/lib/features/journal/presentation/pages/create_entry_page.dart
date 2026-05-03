import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/networking/api_client.dart';
import '../../data/api/media_api.dart' show MediaApi;
import '../providers/journal_provider.dart';

class CreateEntryPage extends ConsumerStatefulWidget {
  const CreateEntryPage({super.key});

  @override
  ConsumerState<CreateEntryPage> createState() => _CreateEntryPageState();
}

class _CreateEntryPageState extends ConsumerState<CreateEntryPage> {
  final _titleCtrl = TextEditingController();
  final _bodyCtrl = TextEditingController();
  String? _selectedMood;
  bool _loading = false;
  String? _error;
  final List<File> _photos = [];
  final List<String> _uploadedMediaIds = [];
  final _picker = ImagePicker();

  final _moods = const [
    ('happy', '😊', 'Радостно'),
    ('excited', '🎉', 'Восторг'),
    ('calm', '😌', 'Спокойно'),
    ('tired', '😴', 'Устал'),
    ('sad', '😢', 'Грустно'),
  ];

  @override
  void dispose() {
    _titleCtrl.dispose();
    _bodyCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickPhoto(ImageSource source) async {
    final picked = await _picker.pickImage(source: source, imageQuality: 85);
    if (picked == null) return;
    setState(() => _photos.add(File(picked.path)));
  }

  void _showPickerSheet() {
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.camera_alt_outlined,
                  color: AppColors.primary),
              title: const Text('Камера'),
              onTap: () {
                Navigator.pop(context);
                _pickPhoto(ImageSource.camera);
              },
            ),
            ListTile(
              leading: const Icon(Icons.photo_library_outlined,
                  color: AppColors.primary),
              title: const Text('Галерея'),
              onTap: () {
                Navigator.pop(context);
                _pickPhoto(ImageSource.gallery);
              },
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _submit() async {
    if (_bodyCtrl.text.trim().isEmpty) {
      setState(() => _error = 'Напишите что-нибудь');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      // Photos are best-effort — failure doesn't block saving the entry.
      final mediaApi = MediaApi(ref.read(apiClientProvider).dio);
      final mediaIds = <String>[];
      int photoFailures = 0;
      for (final photo in _photos) {
        try {
          final id = await mediaApi.uploadFile(photo, purpose: 'photo');
          mediaIds.add(id);
        } catch (e) {
          photoFailures++;
          debugPrint('Photo upload failed: $e');
        }
      }
      _uploadedMediaIds.addAll(mediaIds);

      // Create entry
      final entry = await ref.read(journalApiProvider).createEntry(
            title:
                _titleCtrl.text.trim().isEmpty ? null : _titleCtrl.text.trim(),
            body: _bodyCtrl.text.trim(),
            mood: _selectedMood,
          );

      // Attach media — also best-effort.
      for (var i = 0; i < mediaIds.length; i++) {
        try {
          await mediaApi.attachMedia(entry.id, mediaIds[i], i);
        } catch (e) {
          debugPrint('Attach media failed: $e');
        }
      }

      if (mounted) {
        if (photoFailures > 0) {
          ScaffoldMessenger.of(context).showSnackBar(SnackBar(
            content: Text(
                'Запись сохранена, но не удалось загрузить фото ($photoFailures)'),
          ));
        }
        Navigator.pop(context, true);
      }
    } catch (e, st) {
      String msg;
      if (e is DioException) {
        final code = e.response?.statusCode;
        final body = e.response?.data?.toString() ?? e.message ?? '';
        final url = e.requestOptions.uri.toString();
        msg = 'HTTP ${code ?? '?'} · $url\n$body';
      } else {
        msg = '$e';
      }
      debugPrint('CreateEntry failed: $msg\n$st');
      setState(() => _error = msg);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Новая запись'),
        actions: [
          TextButton(
            onPressed: _loading ? null : _submit,
            child: const Text('Сохранить',
                style: TextStyle(color: AppColors.primary)),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TextField(
              controller: _titleCtrl,
              decoration: const InputDecoration(
                hintText: 'Заголовок (необязательно)',
                border: InputBorder.none,
              ),
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
            ),
            const Divider(),
            TextField(
              controller: _bodyCtrl,
              decoration: const InputDecoration(
                hintText: 'Расскажи о своём путешествии...',
                border: InputBorder.none,
              ),
              maxLines: null,
              minLines: 8,
            ),
            const SizedBox(height: AppSpacing.md),
            // Фото
            Row(
              children: [
                const Text('Фото',
                    style:
                        TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                const Spacer(),
                TextButton.icon(
                  onPressed: _showPickerSheet,
                  icon:
                      const Icon(Icons.add_photo_alternate_outlined, size: 18),
                  label: const Text('Добавить'),
                  style:
                      TextButton.styleFrom(foregroundColor: AppColors.primary),
                ),
              ],
            ),
            if (_photos.isNotEmpty) ...[
              const SizedBox(height: AppSpacing.sm),
              SizedBox(
                height: 88,
                child: ListView.separated(
                  scrollDirection: Axis.horizontal,
                  itemCount: _photos.length + 1,
                  separatorBuilder: (_, __) => const SizedBox(width: 8),
                  itemBuilder: (_, i) {
                    if (i == _photos.length) {
                      return GestureDetector(
                        onTap: _showPickerSheet,
                        child: Container(
                          width: 88,
                          height: 88,
                          decoration: BoxDecoration(
                            border: Border.all(color: AppColors.border),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.add,
                              color: AppColors.textSecondary),
                        ),
                      );
                    }
                    return Stack(
                      children: [
                        ClipRRect(
                          borderRadius: BorderRadius.circular(10),
                          child: Image.file(_photos[i],
                              width: 88, height: 88, fit: BoxFit.cover),
                        ),
                        Positioned(
                          top: 4,
                          right: 4,
                          child: GestureDetector(
                            onTap: () => setState(() => _photos.removeAt(i)),
                            child: Container(
                              width: 22,
                              height: 22,
                              decoration: const BoxDecoration(
                                color: Colors.black54,
                                shape: BoxShape.circle,
                              ),
                              child: const Icon(Icons.close,
                                  size: 14, color: Colors.white),
                            ),
                          ),
                        ),
                      ],
                    );
                  },
                ),
              ),
            ],
            const SizedBox(height: AppSpacing.md),
            // Настроение
            const Text('Настроение',
                style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
            const SizedBox(height: AppSpacing.sm),
            Wrap(
              spacing: 8,
              children: _moods
                  .map((m) => GestureDetector(
                        onTap: () => setState(() => _selectedMood =
                            _selectedMood == m.$1 ? null : m.$1),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 12, vertical: 6),
                          decoration: BoxDecoration(
                            color: _selectedMood == m.$1
                                ? AppColors.primary
                                : Colors.white,
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(color: AppColors.border),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text(m.$2),
                              const SizedBox(width: 4),
                              Text(
                                m.$3,
                                style: TextStyle(
                                  fontSize: 13,
                                  color: _selectedMood == m.$1
                                      ? Colors.white
                                      : AppColors.textPrimary,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ))
                  .toList(),
            ),
            if (_error != null) ...[
              const SizedBox(height: AppSpacing.md),
              Text(_error!, style: const TextStyle(color: AppColors.error)),
            ],
            const SizedBox(height: AppSpacing.xl),
            if (_loading)
              const Center(
                  child: CircularProgressIndicator(color: AppColors.primary)),
          ],
        ),
      ),
    );
  }
}
