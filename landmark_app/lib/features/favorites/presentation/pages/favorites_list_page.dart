import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/widgets/skeleton.dart';
import '../../../map/domain/entities/place.dart';
import '../../../trips/domain/entities/trip.dart';
import '../../../trips/presentation/pages/trip_detail_page.dart';
import '../../../journal/domain/entities/entry.dart';
import '../../../journal/presentation/pages/entry_detail_page.dart';
import '../providers/favorites_provider.dart';
import 'favorites_map_page.dart';

class FavoritesListPage extends ConsumerStatefulWidget {
  const FavoritesListPage({super.key});

  @override
  ConsumerState<FavoritesListPage> createState() => _FavoritesListPageState();
}

class _FavoritesListPageState extends ConsumerState<FavoritesListPage>
    with SingleTickerProviderStateMixin {
  late final TabController _tabs;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Избранное'),
        actions: [
          IconButton(
            icon: const Icon(Icons.map_outlined),
            tooltip: 'На карте',
            onPressed: () => Navigator.push<void>(
              context,
              MaterialPageRoute(builder: (_) => const FavoritesMapPage()),
            ),
          ),
        ],
        bottom: TabBar(
          controller: _tabs,
          labelColor: AppColors.primary,
          unselectedLabelColor: AppColors.textSecondary,
          indicatorColor: AppColors.primary,
          tabs: const [
            Tab(icon: Icon(Icons.place_outlined), text: 'Места'),
            Tab(icon: Icon(Icons.luggage_outlined), text: 'Поездки'),
            Tab(icon: Icon(Icons.book_outlined), text: 'Записи'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabs,
        children: const [
          _PlacesTab(),
          _TripsTab(),
          _EntriesTab(),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Places tab
// ---------------------------------------------------------------------------

class _PlacesTab extends ConsumerWidget {
  const _PlacesTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(favoritePlacesProvider);
    return async.when(
      loading: () => const SkeletonListView(),
      error: (e, _) => _ErrorView(message: '$e', onRetry: () => ref.invalidate(favoritePlacesProvider)),
      data: (places) => places.isEmpty
          ? const _EmptyFavorites(label: 'Нет избранных мест')
          : ListView.separated(
              padding: const EdgeInsets.all(AppSpacing.md),
              itemCount: places.length,
              separatorBuilder: (_, __) => const SizedBox(height: AppSpacing.sm),
              itemBuilder: (_, i) => _PlaceCard(place: places[i]),
            ),
    );
  }
}

class _PlaceCard extends StatelessWidget {
  final Place place;
  const _PlaceCard({required this.place});

  @override
  Widget build(BuildContext context) {
    return Card(
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.all(AppRadius.md)),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        leading: Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: AppColors.primary.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(12),
          ),
          child: const Icon(Icons.place, color: AppColors.primary),
        ),
        title: Text(place.title, style: AppTypography.h3, maxLines: 1, overflow: TextOverflow.ellipsis),
        subtitle: place.city != null
            ? Text('${place.city}${place.country != null ? ', ${place.country}' : ''}',
                style: AppTypography.caption)
            : null,
        trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
        onTap: () {
          // Navigate to map and pan to place
          // For now show a simple detail sheet
          _showPlaceSheet(context, place);
        },
      ),
    );
  }

  void _showPlaceSheet(BuildContext context, Place place) {
    showModalBottomSheet<void>(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(place.title, style: AppTypography.h2),
            if (place.city != null) ...[
              const SizedBox(height: AppSpacing.xs),
              Text(
                '${place.city}${place.country != null ? ', ${place.country}' : ''}',
                style: AppTypography.bodySmall,
              ),
            ],
            if (place.description != null) ...[
              const SizedBox(height: AppSpacing.md),
              Text(place.description!, style: AppTypography.body),
            ],
            const SizedBox(height: AppSpacing.lg),
            Text(
              'Координаты: ${place.latitude.toStringAsFixed(5)}, ${place.longitude.toStringAsFixed(5)}',
              style: AppTypography.caption,
            ),
            const SizedBox(height: AppSpacing.md),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Trips tab
// ---------------------------------------------------------------------------

class _TripsTab extends ConsumerWidget {
  const _TripsTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(favoriteTripsProvider);
    return async.when(
      loading: () => const SkeletonListView(),
      error: (e, _) => _ErrorView(message: '$e', onRetry: () => ref.invalidate(favoriteTripsProvider)),
      data: (trips) => trips.isEmpty
          ? const _EmptyFavorites(label: 'Нет избранных поездок')
          : ListView.separated(
              padding: const EdgeInsets.all(AppSpacing.md),
              itemCount: trips.length,
              separatorBuilder: (_, __) => const SizedBox(height: AppSpacing.sm),
              itemBuilder: (_, i) => _TripCard(trip: trips[i]),
            ),
    );
  }
}

class _TripCard extends StatelessWidget {
  final Trip trip;
  const _TripCard({required this.trip});

  @override
  Widget build(BuildContext context) {
    final fmt = DateFormat('d MMM yyyy', 'ru_RU');
    return Card(
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.all(AppRadius.md)),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        leading: Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: AppColors.accent.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(12),
          ),
          child: const Icon(Icons.luggage, color: AppColors.accent),
        ),
        title: Text(trip.title, style: AppTypography.h3, maxLines: 1, overflow: TextOverflow.ellipsis),
        subtitle: Text(fmt.format(trip.createdAt), style: AppTypography.caption),
        trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
        onTap: () => Navigator.push<void>(
          context,
          MaterialPageRoute(builder: (_) => TripDetailPage(trip: trip)),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Entries tab
// ---------------------------------------------------------------------------

class _EntriesTab extends ConsumerWidget {
  const _EntriesTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(favoriteEntriesProvider);
    return async.when(
      loading: () => const SkeletonListView(),
      error: (e, _) => _ErrorView(message: '$e', onRetry: () => ref.invalidate(favoriteEntriesProvider)),
      data: (entries) => entries.isEmpty
          ? const _EmptyFavorites(label: 'Нет избранных записей')
          : ListView.separated(
              padding: const EdgeInsets.all(AppSpacing.md),
              itemCount: entries.length,
              separatorBuilder: (_, __) => const SizedBox(height: AppSpacing.sm),
              itemBuilder: (_, i) => _EntryCard(entry: entries[i]),
            ),
    );
  }
}

class _EntryCard extends StatelessWidget {
  final JournalEntry entry;
  const _EntryCard({required this.entry});

  @override
  Widget build(BuildContext context) {
    final fmt = DateFormat('d MMM yyyy', 'ru_RU');
    return Card(
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.all(AppRadius.md)),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        leading: Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: Colors.purple.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(12),
          ),
          child: const Icon(Icons.book, color: Colors.purple),
        ),
        title: Text(
          entry.title ?? fmt.format(entry.occurredAt),
          style: AppTypography.h3,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Text(entry.body, style: AppTypography.caption, maxLines: 2, overflow: TextOverflow.ellipsis),
        trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
        onTap: () => Navigator.push<void>(
          context,
          MaterialPageRoute(builder: (_) => EntryDetailPage(entry: entry)),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

class _EmptyFavorites extends StatelessWidget {
  final String label;
  const _EmptyFavorites({required this.label});

  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        const Icon(Icons.favorite_border, size: 64, color: AppColors.textSecondary),
        const SizedBox(height: 16),
        Text(label, style: const TextStyle(fontSize: 18, color: AppColors.textSecondary)),
        const SizedBox(height: 8),
        const Text('Добавляйте места, поездки и записи в избранное',
            style: TextStyle(color: AppColors.textSecondary), textAlign: TextAlign.center),
      ],
    ),
  );
}

class _ErrorView extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;
  const _ErrorView({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        const Icon(Icons.error_outline, size: 48, color: AppColors.textSecondary),
        const SizedBox(height: 12),
        const Text('Ошибка загрузки', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
        const SizedBox(height: 8),
        TextButton(onPressed: onRetry, child: const Text('Повторить')),
      ],
    ),
  );
}
