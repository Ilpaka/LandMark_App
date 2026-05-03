import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/widgets/skeleton.dart';
import '../../domain/entities/trip.dart';
import '../providers/trips_provider.dart';
import 'trip_detail_page.dart';

class PublicTripsPage extends ConsumerWidget {
  const PublicTripsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tripsAsync = ref.watch(publicTripsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Лента поездок'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: tripsAsync.when(
        loading: () => const SkeletonListView(),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.cloud_off_outlined,
                  size: 48, color: AppColors.textSecondary),
              const SizedBox(height: 12),
              const Text('Ошибка загрузки',
                  style: TextStyle(color: AppColors.textSecondary)),
              const SizedBox(height: 12),
              TextButton(
                onPressed: () => ref.invalidate(publicTripsProvider),
                child: const Text('Повторить'),
              ),
            ],
          ),
        ),
        data: (trips) => trips.isEmpty
            ? const Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.explore_outlined,
                        size: 64, color: AppColors.textSecondary),
                    SizedBox(height: 16),
                    Text('Нет публичных поездок',
                        style: TextStyle(
                            fontSize: 18, color: AppColors.textSecondary)),
                  ],
                ),
              )
            : RefreshIndicator(
                onRefresh: () async => ref.invalidate(publicTripsProvider),
                child: ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: trips.length,
                  itemBuilder: (_, i) => _PublicTripCard(trip: trips[i]),
                ),
              ),
      ),
    );
  }
}

class _PublicTripCard extends StatelessWidget {
  final Trip trip;
  const _PublicTripCard({required this.trip});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => Navigator.push(
          context,
          MaterialPageRoute(builder: (_) => TripDetailPage(trip: trip)),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(trip.title,
                        style: const TextStyle(
                            fontSize: 16, fontWeight: FontWeight.w600)),
                  ),
                  _StatusBadge(status: trip.status),
                ],
              ),
              if (trip.subtitle != null && trip.subtitle!.isNotEmpty) ...[
                const SizedBox(height: 6),
                Text(trip.subtitle!,
                    style: const TextStyle(
                        fontSize: 13, color: AppColors.textSecondary),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis),
              ],
              const SizedBox(height: 10),
              Row(
                children: [
                  _MetaItem(
                      icon: Icons.place_outlined,
                      text: '${trip.stopsCount} точек'),
                  const SizedBox(width: 16),
                  _MetaItem(
                      icon: Icons.book_outlined,
                      text: '${trip.entriesCount} записей'),
                  const SizedBox(width: 16),
                  _MetaItem(
                      icon: Icons.photo_outlined,
                      text: '${trip.photosCount} фото'),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatusBadge extends StatelessWidget {
  final String status;
  const _StatusBadge({required this.status});

  Color get _color => switch (status) {
        'planned' => AppColors.warning,
        'in_progress' => AppColors.primary,
        'completed' => AppColors.success,
        _ => AppColors.textSecondary,
      };

  String get _label => switch (status) {
        'planned' => 'Запланировано',
        'in_progress' => 'В пути',
        'completed' => 'Завершено',
        _ => status,
      };

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
        decoration: BoxDecoration(
          color: _color.withValues(alpha: 0.12),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Text(_label,
            style: TextStyle(
                fontSize: 11, color: _color, fontWeight: FontWeight.w600)),
      );
}

class _MetaItem extends StatelessWidget {
  final IconData icon;
  final String text;
  const _MetaItem({required this.icon, required this.text});

  @override
  Widget build(BuildContext context) => Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: AppColors.textSecondary),
          const SizedBox(width: 4),
          Text(text,
              style: const TextStyle(
                  fontSize: 12, color: AppColors.textSecondary)),
        ],
      );
}
