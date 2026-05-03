import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/networking/api_client.dart';

class SuggestPlacePage extends ConsumerStatefulWidget {
  const SuggestPlacePage({super.key});

  @override
  ConsumerState<SuggestPlacePage> createState() => _SuggestPlacePageState();
}

class _SuggestPlacePageState extends ConsumerState<SuggestPlacePage> {
  final _formKey = GlobalKey<FormState>();
  final _titleCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _latCtrl = TextEditingController();
  final _lngCtrl = TextEditingController();
  final _cityCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _titleCtrl.dispose();
    _descCtrl.dispose();
    _latCtrl.dispose();
    _lngCtrl.dispose();
    _cityCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _loading = true);
    try {
      final dio = ref.read(apiClientProvider).dio;

      // 1. Create draft
      final createResp = await dio.post('/v1/places', data: {
        'title': _titleCtrl.text.trim(),
        'description':
            _descCtrl.text.trim().isEmpty ? null : _descCtrl.text.trim(),
        'latitude': double.parse(_latCtrl.text.trim()),
        'longitude': double.parse(_lngCtrl.text.trim()),
        'city': _cityCtrl.text.trim().isEmpty ? null : _cityCtrl.text.trim(),
      });
      final placeId = (createResp.data as Map<String, dynamic>)['id'] as String;

      // 2. Submit for moderation
      await dio.post('/v1/places/$placeId/submit');

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Место отправлено на модерацию!')),
        );
        Navigator.of(context).pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Предложить место'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _field(_titleCtrl, 'Название *', required: true),
            const SizedBox(height: 12),
            _field(_descCtrl, 'Описание', maxLines: 4),
            const SizedBox(height: 12),
            Row(children: [
              Expanded(
                  child: _field(_latCtrl, 'Широта *',
                      required: true, keyboard: TextInputType.number)),
              const SizedBox(width: 12),
              Expanded(
                  child: _field(_lngCtrl, 'Долгота *',
                      required: true, keyboard: TextInputType.number)),
            ]),
            const SizedBox(height: 12),
            _field(_cityCtrl, 'Город'),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: _loading ? null : _submit,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                minimumSize: const Size.fromHeight(48),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12)),
              ),
              child: _loading
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(
                          color: Colors.white, strokeWidth: 2))
                  : const Text('Отправить на проверку'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _field(
    TextEditingController ctrl,
    String label, {
    bool required = false,
    int maxLines = 1,
    TextInputType? keyboard,
  }) {
    return TextFormField(
      controller: ctrl,
      maxLines: maxLines,
      keyboardType: keyboard,
      decoration: InputDecoration(
        labelText: label,
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
        filled: true,
        fillColor: AppColors.surface,
      ),
      validator: required
          ? (v) => (v == null || v.trim().isEmpty) ? 'Обязательное поле' : null
          : null,
    );
  }
}
