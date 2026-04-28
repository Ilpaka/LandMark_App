import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
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
      await ref.read(journalApiProvider).createEntry(
        title: _titleCtrl.text.trim().isEmpty ? null : _titleCtrl.text.trim(),
        body: _bodyCtrl.text.trim(),
        mood: _selectedMood,
      );
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      setState(() => _error = 'Не удалось сохранить запись');
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
            child: const Text('Сохранить', style: TextStyle(color: AppColors.primary)),
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
            // Выбор настроения
            const Text(
              'Настроение',
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
            ),
            const SizedBox(height: AppSpacing.sm),
            Wrap(
              spacing: 8,
              children: _moods.map((m) => GestureDetector(
                onTap: () => setState(
                  () => _selectedMood = _selectedMood == m.$1 ? null : m.$1,
                ),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: _selectedMood == m.$1 ? AppColors.primary : Colors.white,
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
              )).toList(),
            ),
            if (_error != null) ...[
              const SizedBox(height: AppSpacing.md),
              Text(_error!, style: const TextStyle(color: AppColors.error)),
            ],
            const SizedBox(height: AppSpacing.xl),
            if (_loading)
              const Center(child: CircularProgressIndicator(color: AppColors.primary)),
          ],
        ),
      ),
    );
  }
}
