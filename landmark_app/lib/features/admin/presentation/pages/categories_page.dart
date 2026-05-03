import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/networking/api_client.dart';
import '../../../map/data/api/places_api.dart';
import '../../../map/domain/entities/place.dart';

final _categoriesProvider = FutureProvider<List<PlaceCategory>>((ref) async {
  final api = PlacesApi(ref.read(apiClientProvider).dio);
  return api.getCategories();
});

class CategoriesPage extends ConsumerWidget {
  const CategoriesPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final catsAsync = ref.watch(_categoriesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Категории мест'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () => _showEditDialog(context, ref, null),
          ),
        ],
      ),
      body: catsAsync.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('Ошибка: $e', textAlign: TextAlign.center),
              const SizedBox(height: 12),
              ElevatedButton(
                onPressed: () => ref.invalidate(_categoriesProvider),
                child: const Text('Повторить'),
              ),
            ],
          ),
        ),
        data: (cats) => cats.isEmpty
            ? const Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.category_outlined,
                        size: 64, color: AppColors.textSecondary),
                    SizedBox(height: 12),
                    Text('Нет категорий',
                        style: TextStyle(color: AppColors.textSecondary)),
                  ],
                ),
              )
            : RefreshIndicator(
                color: AppColors.primary,
                onRefresh: () async => ref.invalidate(_categoriesProvider),
                child: ListView.separated(
                  padding: const EdgeInsets.all(16),
                  itemCount: cats.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 8),
                  itemBuilder: (_, i) => _CategoryTile(
                    category: cats[i],
                    onEdit: () => _showEditDialog(context, ref, cats[i]),
                    onDelete: () => _confirmDelete(context, ref, cats[i]),
                  ),
                ),
              ),
      ),
    );
  }

  void _showEditDialog(
      BuildContext context, WidgetRef ref, PlaceCategory? existing) {
    showDialog<void>(
      context: context,
      builder: (ctx) => _CategoryDialog(
        existing: existing,
        onSaved: () {
          ref.invalidate(_categoriesProvider);
          Navigator.pop(ctx);
        },
      ),
    );
  }

  void _confirmDelete(BuildContext context, WidgetRef ref, PlaceCategory cat) {
    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить категорию?'),
        content: Text('«${cat.title}» будет деактивирована.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx), child: const Text('Отмена')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red, foregroundColor: Colors.white),
            onPressed: () async {
              Navigator.pop(ctx);
              try {
                final dio = ref.read(apiClientProvider).dio;
                await dio.delete('/v1/places/categories/${cat.id}');
                ref.invalidate(_categoriesProvider);
              } catch (e) {
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Ошибка: $e')),
                  );
                }
              }
            },
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
  }
}

class _CategoryTile extends StatelessWidget {
  final PlaceCategory category;
  final VoidCallback onEdit;
  final VoidCallback onDelete;
  const _CategoryTile(
      {required this.category, required this.onEdit, required this.onDelete});

  @override
  Widget build(BuildContext context) {
    final color = _parseColor(category.color);
    return Card(
      elevation: 0,
      color: AppColors.surface,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        leading: Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(10),
          ),
          alignment: Alignment.center,
          child: Text(category.icon, style: const TextStyle(fontSize: 20)),
        ),
        title: Text(category.title,
            style: const TextStyle(fontWeight: FontWeight.w500)),
        subtitle: Text('/${category.slug}',
            style:
                const TextStyle(fontSize: 12, color: AppColors.textSecondary)),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            IconButton(
              icon: const Icon(Icons.edit_outlined, size: 20),
              color: AppColors.primary,
              onPressed: onEdit,
            ),
            IconButton(
              icon: const Icon(Icons.delete_outline, size: 20),
              color: Colors.red,
              onPressed: onDelete,
            ),
          ],
        ),
      ),
    );
  }

  Color _parseColor(String hex) {
    try {
      final clean = hex.replaceFirst('#', '');
      return Color(int.parse('FF$clean', radix: 16));
    } catch (_) {
      return AppColors.primary;
    }
  }
}

class _CategoryDialog extends ConsumerStatefulWidget {
  final PlaceCategory? existing;
  final VoidCallback onSaved;
  const _CategoryDialog({this.existing, required this.onSaved});

  @override
  ConsumerState<_CategoryDialog> createState() => _CategoryDialogState();
}

class _CategoryDialogState extends ConsumerState<_CategoryDialog> {
  late final TextEditingController _slugCtrl;
  late final TextEditingController _titleCtrl;
  late final TextEditingController _iconCtrl;
  late final TextEditingController _colorCtrl;
  late final TextEditingController _orderCtrl;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final e = widget.existing;
    _slugCtrl = TextEditingController(text: e?.slug ?? '');
    _titleCtrl = TextEditingController(text: e?.title ?? '');
    _iconCtrl = TextEditingController(text: e?.icon ?? '🏷️');
    _colorCtrl = TextEditingController(text: e?.color ?? '#4A90D9');
    _orderCtrl = TextEditingController(text: '0');
  }

  @override
  void dispose() {
    _slugCtrl.dispose();
    _titleCtrl.dispose();
    _iconCtrl.dispose();
    _colorCtrl.dispose();
    _orderCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (_titleCtrl.text.trim().isEmpty) return;
    setState(() => _saving = true);
    try {
      final dio = ref.read(apiClientProvider).dio;
      final body = {
        'title': _titleCtrl.text.trim(),
        'icon': _iconCtrl.text.trim(),
        'color': _colorCtrl.text.trim(),
        'sort_order': int.tryParse(_orderCtrl.text) ?? 0,
      };
      if (widget.existing == null) {
        body['slug'] = _slugCtrl.text.trim();
        await dio.post('/v1/places/categories', data: body);
      } else {
        await dio.patch('/v1/places/categories/${widget.existing!.id}',
            data: body);
      }
      widget.onSaved();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(
          widget.existing == null ? 'Новая категория' : 'Изменить категорию'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (widget.existing == null)
              TextField(
                controller: _slugCtrl,
                decoration:
                    const InputDecoration(labelText: 'Slug (en, без пробелов)'),
              ),
            TextField(
              controller: _titleCtrl,
              decoration: const InputDecoration(labelText: 'Название *'),
            ),
            TextField(
              controller: _iconCtrl,
              decoration: const InputDecoration(labelText: 'Иконка (эмодзи)'),
            ),
            TextField(
              controller: _colorCtrl,
              decoration: const InputDecoration(labelText: 'Цвет (#hex)'),
            ),
            TextField(
              controller: _orderCtrl,
              decoration:
                  const InputDecoration(labelText: 'Порядок сортировки'),
              keyboardType: TextInputType.number,
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Отмена'),
        ),
        ElevatedButton(
          style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primary,
              foregroundColor: Colors.white),
          onPressed: _saving ? null : _save,
          child: _saving
              ? const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                      strokeWidth: 2, color: Colors.white))
              : const Text('Сохранить'),
        ),
      ],
    );
  }
}
