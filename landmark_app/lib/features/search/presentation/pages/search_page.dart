import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/networking/api_client.dart';
import '../../../map/domain/entities/place.dart';
import '../../../map/presentation/pages/place_detail_page.dart';
import '../../../trips/domain/entities/trip.dart';
import '../../../trips/presentation/pages/trip_detail_page.dart';
import '../../../journal/domain/entities/entry.dart';
import '../../../journal/presentation/pages/entry_detail_page.dart';

class SearchPage extends ConsumerStatefulWidget {
  const SearchPage({super.key});

  @override
  ConsumerState<SearchPage> createState() => _SearchPageState();
}

class _SearchPageState extends ConsumerState<SearchPage>
    with SingleTickerProviderStateMixin {
  late final TabController _tabCtrl;
  final _ctrl = TextEditingController();
  Timer? _debounce;
  bool _loading = false;

  List<Place> _places = [];
  List<Trip> _trips = [];
  List<JournalEntry> _entries = [];
  String _lastQuery = '';

  @override
  void initState() {
    super.initState();
    _tabCtrl = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _ctrl.dispose();
    _debounce?.cancel();
    _tabCtrl.dispose();
    super.dispose();
  }

  void _onChanged(String q) {
    _debounce?.cancel();
    if (q.trim().isEmpty) {
      setState(() {
        _places = [];
        _trips = [];
        _entries = [];
        _loading = false;
      });
      return;
    }
    setState(() => _loading = true);
    _debounce = Timer(const Duration(milliseconds: 400), () => _search(q.trim()));
  }

  Future<void> _search(String q) async {
    if (q == _lastQuery) return;
    _lastQuery = q;
    final dio = ref.read(apiClientProvider).dio;
    try {
      final results = await Future.wait([
        dio.get('/v1/places', queryParameters: {'q': q, 'limit': 20}).then((r) {
          final data = r.data as Map<String, dynamic>;
          return (data['places'] as List<dynamic>? ?? [])
              .map((e) => Place.fromJson(e as Map<String, dynamic>))
              .toList();
        }).catchError((_) => <Place>[]),
        dio.get('/v1/trips', queryParameters: {'q': q, 'limit': 20}).then((r) {
          final data = r.data as Map<String, dynamic>;
          return (data['trips'] as List<dynamic>? ?? [])
              .map((e) => Trip.fromJson(e as Map<String, dynamic>))
              .toList();
        }).catchError((_) => <Trip>[]),
        dio.get('/v1/journal/entries', queryParameters: {'q': q, 'limit': 20}).then((r) {
          final data = r.data as Map<String, dynamic>;
          return (data['entries'] as List<dynamic>? ?? [])
              .map((e) => JournalEntry.fromJson(e as Map<String, dynamic>))
              .toList();
        }).catchError((_) => <JournalEntry>[]),
      ]);

      if (mounted) {
        setState(() {
          _places = results[0] as List<Place>;
          _trips = results[1] as List<Trip>;
          _entries = results[2] as List<JournalEntry>;
          _loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final hasQuery = _ctrl.text.isNotEmpty;
    return Scaffold(
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        title: TextField(
          controller: _ctrl,
          autofocus: true,
          decoration: const InputDecoration(
            hintText: 'Поиск...',
            border: InputBorder.none,
            hintStyle: TextStyle(color: AppColors.textSecondary),
          ),
          onChanged: _onChanged,
        ),
        actions: [
          if (hasQuery)
            IconButton(
              icon: const Icon(Icons.clear),
              onPressed: () {
                _ctrl.clear();
                setState(() {
                  _places = [];
                  _trips = [];
                  _entries = [];
                  _loading = false;
                  _lastQuery = '';
                });
              },
            ),
        ],
        bottom: TabBar(
          controller: _tabCtrl,
          labelColor: AppColors.primary,
          unselectedLabelColor: AppColors.textSecondary,
          indicatorColor: AppColors.primary,
          tabs: [
            Tab(text: 'Места${_places.isNotEmpty ? " (${_places.length})" : ""}'),
            Tab(text: 'Поездки${_trips.isNotEmpty ? " (${_trips.length})" : ""}'),
            Tab(text: 'Записи${_entries.isNotEmpty ? " (${_entries.length})" : ""}'),
          ],
        ),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator(color: AppColors.primary))
          : TabBarView(
              controller: _tabCtrl,
              children: [
                _PlacesTab(places: _places, hasQuery: hasQuery),
                _TripsTab(trips: _trips, hasQuery: hasQuery),
                _EntriesTab(entries: _entries, hasQuery: hasQuery),
              ],
            ),
    );
  }
}

// ---------------------------------------------------------------------------

class _PlacesTab extends StatelessWidget {
  final List<Place> places;
  final bool hasQuery;
  const _PlacesTab({required this.places, required this.hasQuery});

  @override
  Widget build(BuildContext context) {
    if (!hasQuery) return const _EmptyHint(hint: 'Введите название места');
    if (places.isEmpty) return const _NoResults();
    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: places.length,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (_, i) {
        final p = places[i];
        return ListTile(
          leading: Container(
            width: 40, height: 40,
            decoration: BoxDecoration(
              color: AppColors.primary.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(Icons.place, color: AppColors.primary, size: 20),
          ),
          title: Text(p.title, style: const TextStyle(fontWeight: FontWeight.w600)),
          subtitle: p.city != null
              ? Text(p.city!, style: const TextStyle(fontSize: 12, color: AppColors.textSecondary))
              : null,
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => PlaceDetailPage(place: p)),
          ),
        );
      },
    );
  }
}

class _TripsTab extends StatelessWidget {
  final List<Trip> trips;
  final bool hasQuery;
  const _TripsTab({required this.trips, required this.hasQuery});

  @override
  Widget build(BuildContext context) {
    if (!hasQuery) return const _EmptyHint(hint: 'Введите название поездки');
    if (trips.isEmpty) return const _NoResults();
    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: trips.length,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (_, i) {
        final t = trips[i];
        return ListTile(
          leading: Container(
            width: 40, height: 40,
            decoration: BoxDecoration(
              color: AppColors.accent.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(Icons.luggage, color: AppColors.accent, size: 20),
          ),
          title: Text(t.title, style: const TextStyle(fontWeight: FontWeight.w600)),
          subtitle: Text('${t.stopsCount} точек · ${t.status}',
              style: const TextStyle(fontSize: 12, color: AppColors.textSecondary)),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => TripDetailPage(trip: t)),
          ),
        );
      },
    );
  }
}

class _EntriesTab extends StatelessWidget {
  final List<JournalEntry> entries;
  final bool hasQuery;
  const _EntriesTab({required this.entries, required this.hasQuery});

  @override
  Widget build(BuildContext context) {
    if (!hasQuery) return const _EmptyHint(hint: 'Введите текст записи');
    if (entries.isEmpty) return const _NoResults();
    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: entries.length,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (_, i) {
        final e = entries[i];
        return ListTile(
          leading: Container(
            width: 40, height: 40,
            decoration: BoxDecoration(
              color: Colors.teal.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(Icons.book_outlined, color: Colors.teal, size: 20),
          ),
          title: Text(
            e.title?.isNotEmpty == true ? e.title! : e.body,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(fontWeight: FontWeight.w600),
          ),
          subtitle: e.title?.isNotEmpty == true
              ? Text(e.body, maxLines: 1, overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontSize: 12, color: AppColors.textSecondary))
              : null,
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => EntryDetailPage(entry: e)),
          ),
        );
      },
    );
  }
}

class _EmptyHint extends StatelessWidget {
  final String hint;
  const _EmptyHint({required this.hint});

  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Icon(Icons.search, size: 56, color: AppColors.textSecondary),
        const SizedBox(height: 12),
        Text(hint, style: const TextStyle(color: AppColors.textSecondary, fontSize: 15)),
      ],
    ),
  );
}

class _NoResults extends StatelessWidget {
  const _NoResults();

  @override
  Widget build(BuildContext context) => const Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.search_off, size: 56, color: AppColors.textSecondary),
        SizedBox(height: 12),
        Text('Ничего не найдено', style: TextStyle(color: AppColors.textSecondary, fontSize: 15)),
      ],
    ),
  );
}
