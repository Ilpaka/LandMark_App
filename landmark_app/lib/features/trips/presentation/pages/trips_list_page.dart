import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/design/components/wl_button.dart';
import '../../../../core/widgets/skeleton.dart';
import '../../domain/entities/trip.dart';
import '../providers/trips_provider.dart';
import 'create_trip_page.dart';
import 'trip_detail_page.dart';

class TripsListPage extends ConsumerWidget {
  const TripsListPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tripsAsync = ref.watch(tripsListProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Поездки'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () => _showCreateDialog(context, ref),
          ),
        ],
      ),
      body: tripsAsync.when(
        data: (trips) => trips.isEmpty
            ? _EmptyTrips(onCreate: () => _showCreateDialog(context, ref))
            : _PaginatedTripList(trips: trips),
        loading: () => const SkeletonListView(),
        error: (e, _) => const Center(
            child: Text('Ошибка загрузки', style: AppTypography.body)),
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref) async {
    final result = await Navigator.push<bool>(
      context,
      MaterialPageRoute(builder: (_) => const CreateTripPage()),
    );
    if (result == true) ref.invalidate(tripsListProvider);
  }
}

class _PaginatedTripList extends ConsumerWidget {
  final List<Trip> trips;
  const _PaginatedTripList({required this.trips});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifier = ref.read(tripsListProvider.notifier);
    return NotificationListener<ScrollNotification>(
      onNotification: (n) {
        if (n is ScrollEndNotification &&
            n.metrics.extentAfter < 200 &&
            notifier.hasMore) {
          notifier.loadMore();
        }
        return false;
      },
      child: ListView.builder(
        padding: const EdgeInsets.all(AppSpacing.md),
        itemCount: trips.length + 1,
        itemBuilder: (_, i) {
          if (i == trips.length) {
            return notifier.hasMore
                ? const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Center(
                        child: CircularProgressIndicator(
                            color: AppColors.primary, strokeWidth: 2)))
                : const SizedBox.shrink();
          }
          return _TripCard(trip: trips[i]);
        },
      ),
    );
  }
}

class _TripCard extends StatelessWidget {
  final Trip trip;
  const _TripCard({required this.trip});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () => Navigator.push(
        context,
        MaterialPageRoute(builder: (_) => TripDetailPage(trip: trip)),
      ),
      borderRadius: const BorderRadius.all(AppRadius.md),
      child: Card(
        margin: const EdgeInsets.only(bottom: AppSpacing.md),
        shape: const RoundedRectangleBorder(
            borderRadius: BorderRadius.all(AppRadius.md)),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(child: Text(trip.title, style: AppTypography.h3)),
                  _StatusChip(status: trip.status),
                ],
              ),
              const SizedBox(height: AppSpacing.sm),
              Row(
                children: [
                  _CountBadge(
                      icon: Icons.place,
                      count: trip.stopsCount,
                      label: 'точек'),
                  const SizedBox(width: AppSpacing.md),
                  _CountBadge(
                      icon: Icons.book,
                      count: trip.entriesCount,
                      label: 'записей'),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatusChip extends StatelessWidget {
  final String status;
  const _StatusChip({required this.status});

  Color get _color => switch (status) {
        'draft' => AppColors.textSecondary,
        'planned' => AppColors.warning,
        'in_progress' => AppColors.primary,
        'completed' => AppColors.success,
        _ => AppColors.textSecondary,
      };

  String get _label => switch (status) {
        'draft' => 'Черновик',
        'planned' => 'Запланировано',
        'in_progress' => 'В поездке',
        'completed' => 'Завершено',
        _ => status,
      };

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
        decoration: BoxDecoration(
          color: _color.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(
          _label,
          style: TextStyle(
              fontSize: 11, color: _color, fontWeight: FontWeight.w500),
        ),
      );
}

class _CountBadge extends StatelessWidget {
  final IconData icon;
  final int count;
  final String label;
  const _CountBadge(
      {required this.icon, required this.count, required this.label});

  @override
  Widget build(BuildContext context) => Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: AppColors.textSecondary),
          const SizedBox(width: 4),
          Text('$count $label', style: AppTypography.caption),
        ],
      );
}

class _EmptyTrips extends StatelessWidget {
  final VoidCallback onCreate;
  const _EmptyTrips({required this.onCreate});

  @override
  Widget build(BuildContext context) => Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.luggage_outlined,
                size: 64, color: AppColors.textSecondary),
            const SizedBox(height: 16),
            const Text('Нет поездок', style: AppTypography.h3),
            const SizedBox(height: 8),
            const Text('Создайте первую поездку',
                style: AppTypography.bodySmall),
            const SizedBox(height: 24),
            SizedBox(
              width: 200,
              child: WlButton(label: 'Создать поездку', onPressed: onCreate),
            ),
          ],
        ),
      );
}
