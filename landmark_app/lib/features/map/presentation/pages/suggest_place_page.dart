import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';

import '../../../../core/design/tokens.dart';
import '../../../../core/networking/api_client.dart';
import '../../../../core/widgets/location_picker_page.dart';

class SuggestPlacePage extends ConsumerStatefulWidget {
  const SuggestPlacePage({super.key});

  @override
  ConsumerState<SuggestPlacePage> createState() => _SuggestPlacePageState();
}

class _SuggestPlacePageState extends ConsumerState<SuggestPlacePage> {
  final _formKey = GlobalKey<FormState>();
  final _titleCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _cityCtrl = TextEditingController();

  LatLng? _coords;
  bool _coordsTouched = false;
  bool _loading = false;
  bool _isPrivate = false; // false = публичная (с модерацией), true = приватная

  @override
  void dispose() {
    _titleCtrl.dispose();
    _descCtrl.dispose();
    _cityCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickCoords() async {
    final picked = await LocationPickerPage.show(
      context,
      initial: _coords,
      title: 'Где находится место?',
    );
    if (picked != null) {
      setState(() {
        _coords = picked;
        _coordsTouched = true;
      });
    } else {
      // Пользователь закрыл пикер — учитываем как факт обращения, чтобы
      // показать ошибку валидации, если координаты так и не выбраны.
      setState(() => _coordsTouched = true);
    }
  }

  Future<void> _submit() async {
    final formOk = _formKey.currentState!.validate();
    setState(() => _coordsTouched = true);
    if (!formOk || _coords == null) return;
    setState(() => _loading = true);
    try {
      final dio = ref.read(apiClientProvider).dio;
      final visibility = _isPrivate ? 'private' : 'public';

      // 1. Создаём место с нужной видимостью.
      // - private → бэкенд сразу публикует, ниже /submit не нужен.
      // - public  → создаётся draft, дальше отправляем на модерацию.
      final createResp = await dio.post('/v1/places', data: {
        'title': _titleCtrl.text.trim(),
        'description':
            _descCtrl.text.trim().isEmpty ? null : _descCtrl.text.trim(),
        'latitude': _coords!.latitude,
        'longitude': _coords!.longitude,
        'city': _cityCtrl.text.trim().isEmpty ? null : _cityCtrl.text.trim(),
        'visibility': visibility,
      });
      final placeId = (createResp.data as Map<String, dynamic>)['id'] as String;

      if (!_isPrivate) {
        // 2. Только публичные идут на модерацию.
        await dio.post('/v1/places/$placeId/submit');
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              _isPrivate
                  ? 'Точка добавлена на вашу карту'
                  : 'Место отправлено на модерацию!',
            ),
          ),
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
            _LocationPickerField(
              coords: _coords,
              showError: _coordsTouched && _coords == null,
              onTap: _pickCoords,
            ),
            const SizedBox(height: 12),
            _field(_cityCtrl, 'Город'),
            const SizedBox(height: 16),
            _VisibilityPicker(
              isPrivate: _isPrivate,
              onChanged: (v) => setState(() => _isPrivate = v),
            ),
            const SizedBox(height: 20),
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
                  : Text(
                      _isPrivate
                          ? 'Добавить на мою карту'
                          : 'Отправить на проверку',
                    ),
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

class _VisibilityPicker extends StatelessWidget {
  const _VisibilityPicker({
    required this.isPrivate,
    required this.onChanged,
  });

  final bool isPrivate;
  final ValueChanged<bool> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Видимость',
            style: TextStyle(
              fontSize: 12,
              color: AppColors.textPrimary.withValues(alpha: 0.7),
            ),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: _VisibilityChip(
                  selected: !isPrivate,
                  icon: Icons.public,
                  title: 'Публичная',
                  subtitle: 'Видна всем после модерации',
                  onTap: () => onChanged(false),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: _VisibilityChip(
                  selected: isPrivate,
                  icon: Icons.lock_outline,
                  title: 'Приватная',
                  subtitle: 'Только для меня, без модерации',
                  onTap: () => onChanged(true),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _VisibilityChip extends StatelessWidget {
  const _VisibilityChip({
    required this.selected,
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
  });

  final bool selected;
  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final borderColor = selected ? AppColors.primary : AppColors.border;
    final bgColor = selected
        ? AppColors.primary.withValues(alpha: 0.08)
        : AppColors.background;
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 10),
        decoration: BoxDecoration(
          color: bgColor,
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: borderColor, width: selected ? 1.5 : 1),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                Icon(
                  icon,
                  size: 16,
                  color: selected
                      ? AppColors.primary
                      : AppColors.textPrimary.withValues(alpha: 0.6),
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    title,
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color:
                          selected ? AppColors.primary : AppColors.textPrimary,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),
            Text(
              subtitle,
              style: TextStyle(
                fontSize: 11,
                color: AppColors.textPrimary.withValues(alpha: 0.65),
                height: 1.25,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _LocationPickerField extends StatelessWidget {
  const _LocationPickerField({
    required this.coords,
    required this.showError,
    required this.onTap,
  });

  final LatLng? coords;
  final bool showError;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final hasValue = coords != null;
    final borderColor = showError
        ? Colors.red
        : hasValue
            ? AppColors.primary
            : AppColors.border;

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: borderColor, width: showError ? 1.5 : 1),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  hasValue
                      ? Icons.location_on
                      : Icons.add_location_alt_outlined,
                  size: 20,
                  color: hasValue
                      ? AppColors.primary
                      : AppColors.textPrimary.withValues(alpha: 0.6),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    hasValue ? 'Точка на карте *' : 'Выбрать точку на карте *',
                    style: TextStyle(
                      fontSize: 12,
                      color: AppColors.textPrimary.withValues(alpha: 0.7),
                    ),
                  ),
                ),
                Icon(
                  hasValue ? Icons.edit : Icons.chevron_right,
                  size: 18,
                  color: AppColors.textPrimary.withValues(alpha: 0.5),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Text(
              hasValue
                  ? '${coords!.latitude.toStringAsFixed(6)}, '
                      '${coords!.longitude.toStringAsFixed(6)}'
                  : 'Координаты не выбраны',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: hasValue
                    ? AppColors.textPrimary
                    : AppColors.textPrimary.withValues(alpha: 0.5),
                fontFeatures: const [FontFeature.tabularFigures()],
              ),
            ),
            if (showError) ...[
              const SizedBox(height: 6),
              const Text(
                'Поставьте метку на карте',
                style: TextStyle(fontSize: 12, color: Colors.red),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
