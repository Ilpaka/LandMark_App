import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/components/wl_button.dart';
import '../providers/trips_provider.dart';

/// TR-02 / TR-03 — двухшаговый визард создания поездки.
class CreateTripPage extends ConsumerStatefulWidget {
  const CreateTripPage({super.key});

  @override
  ConsumerState<CreateTripPage> createState() => _CreateTripPageState();
}

class _CreateTripPageState extends ConsumerState<CreateTripPage> {
  final _titleCtrl = TextEditingController();
  final _subtitleCtrl = TextEditingController();
  DateTime? _startDate;
  DateTime? _endDate;
  bool _loading = false;
  int _step = 0; // 0 = название, 1 = даты

  @override
  void dispose() {
    _titleCtrl.dispose();
    _subtitleCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_step == 0 ? 'Новая поездка' : 'Даты поездки'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            if (_step == 1) {
              setState(() => _step = 0);
            } else {
              Navigator.pop(context);
            }
          },
        ),
      ),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: _step == 0
            ? _StepTitle(
                titleCtrl: _titleCtrl,
                subtitleCtrl: _subtitleCtrl,
                onNext: () {
                  if (_titleCtrl.text.trim().isEmpty) return;
                  setState(() => _step = 1);
                },
              )
            : _StepDates(
                startDate: _startDate,
                endDate: _endDate,
                onStartPick: () => _pickDate(isStart: true),
                onEndPick: () => _pickDate(isStart: false),
                loading: _loading,
                onCreate: _create,
              ),
      ),
    );
  }

  Future<void> _pickDate({required bool isStart}) async {
    final picked = await showDatePicker(
      context: context,
      initialDate: isStart
          ? (_startDate ?? DateTime.now())
          : (_endDate ??
              (_startDate ?? DateTime.now()).add(const Duration(days: 3))),
      firstDate: DateTime(2020),
      lastDate: DateTime(2035),
    );
    if (picked == null) return;
    setState(() {
      if (isStart) {
        _startDate = picked;
        if (_endDate != null && _endDate!.isBefore(picked)) _endDate = null;
      } else {
        _endDate = picked;
      }
    });
  }

  Future<void> _create() async {
    if (_loading) return;
    setState(() => _loading = true);
    try {
      await ref.read(tripsApiProvider).createTrip(_titleCtrl.text.trim());
      ref.invalidate(tripsListProvider);
      if (mounted) Navigator.pop(context, true);
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
}

// ---------------------------------------------------------------------------
// Шаг 1 — название
// ---------------------------------------------------------------------------

class _StepTitle extends StatelessWidget {
  final TextEditingController titleCtrl;
  final TextEditingController subtitleCtrl;
  final VoidCallback onNext;

  const _StepTitle({
    required this.titleCtrl,
    required this.subtitleCtrl,
    required this.onNext,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Как назовём поездку?',
            style: TextStyle(fontSize: 22, fontWeight: FontWeight.w700)),
        const SizedBox(height: AppSpacing.lg),
        TextField(
          controller: titleCtrl,
          autofocus: true,
          textCapitalization: TextCapitalization.sentences,
          decoration: InputDecoration(
            labelText: 'Название *',
            hintText: 'Например, «Майский Питер»',
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.primary, width: 2),
            ),
          ),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: subtitleCtrl,
          textCapitalization: TextCapitalization.sentences,
          decoration: InputDecoration(
            labelText: 'Описание (необязательно)',
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.primary, width: 2),
            ),
          ),
        ),
        const Spacer(),
        ListenableBuilder(
          listenable: titleCtrl,
          builder: (_, __) => WlButton(
            label: 'Далее',
            onPressed: titleCtrl.text.trim().isNotEmpty ? onNext : null,
          ),
        ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Шаг 2 — даты
// ---------------------------------------------------------------------------

class _StepDates extends StatelessWidget {
  final DateTime? startDate;
  final DateTime? endDate;
  final VoidCallback onStartPick;
  final VoidCallback onEndPick;
  final bool loading;
  final VoidCallback onCreate;

  const _StepDates({
    required this.startDate,
    required this.endDate,
    required this.onStartPick,
    required this.onEndPick,
    required this.loading,
    required this.onCreate,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Когда едем?',
            style: TextStyle(fontSize: 22, fontWeight: FontWeight.w700)),
        const SizedBox(height: 8),
        const Text('Даты можно указать позже',
            style: TextStyle(color: AppColors.textSecondary)),
        const SizedBox(height: AppSpacing.xl),
        _DateRow(
          label: 'Дата отъезда',
          date: startDate,
          icon: Icons.flight_takeoff,
          onTap: onStartPick,
        ),
        const SizedBox(height: AppSpacing.md),
        _DateRow(
          label: 'Дата возвращения',
          date: endDate,
          icon: Icons.flight_land,
          onTap: onEndPick,
        ),
        const Spacer(),
        WlButton(
          label: loading ? 'Создаём...' : 'Создать поездку',
          onPressed: loading ? null : onCreate,
        ),
      ],
    );
  }
}

class _DateRow extends StatelessWidget {
  final String label;
  final DateTime? date;
  final IconData icon;
  final VoidCallback onTap;

  const _DateRow({
    required this.label,
    required this.date,
    required this.icon,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          border: Border.all(
            color: date != null ? AppColors.primary : AppColors.border,
            width: date != null ? 2 : 1,
          ),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(icon,
                color:
                    date != null ? AppColors.primary : AppColors.textSecondary),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(label,
                      style: const TextStyle(
                          fontSize: 12, color: AppColors.textSecondary)),
                  const SizedBox(height: 2),
                  Text(
                    date != null
                        ? '${date!.day.toString().padLeft(2, '0')}.${date!.month.toString().padLeft(2, '0')}.${date!.year}'
                        : 'Не указано',
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: date != null
                          ? AppColors.textPrimary
                          : AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            const Icon(Icons.calendar_today_outlined,
                size: 18, color: AppColors.textSecondary),
          ],
        ),
      ),
    );
  }
}
