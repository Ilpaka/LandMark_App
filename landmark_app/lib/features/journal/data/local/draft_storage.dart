import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

class JournalDraft {
  final String id;
  final String tripId;
  final String title;
  final String body;
  final String? mood;
  final DateTime createdAt;

  const JournalDraft({
    required this.id,
    required this.tripId,
    required this.title,
    required this.body,
    this.mood,
    required this.createdAt,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'trip_id': tripId,
        'title': title,
        'body': body,
        if (mood != null) 'mood': mood,
        'created_at': createdAt.toIso8601String(),
      };

  factory JournalDraft.fromJson(Map<String, dynamic> j) => JournalDraft(
        id: j['id'] as String,
        tripId: j['trip_id'] as String,
        title: j['title'] as String,
        body: j['body'] as String,
        mood: j['mood'] as String?,
        createdAt: DateTime.parse(j['created_at'] as String),
      );
}

const _kDraftsKey = 'journal_drafts';

class DraftStorage {
  final SharedPreferences _prefs;
  DraftStorage(this._prefs);

  List<JournalDraft> getDrafts() {
    final raw = _prefs.getString(_kDraftsKey);
    if (raw == null) return [];
    final list = json.decode(raw) as List<dynamic>;
    return list
        .map((e) => JournalDraft.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> saveDraft(JournalDraft draft) async {
    final drafts = getDrafts();
    final idx = drafts.indexWhere((d) => d.id == draft.id);
    if (idx >= 0) {
      drafts[idx] = draft;
    } else {
      drafts.add(draft);
    }
    await _prefs.setString(
        _kDraftsKey, json.encode(drafts.map((d) => d.toJson()).toList()));
  }

  Future<void> removeDraft(String id) async {
    final drafts = getDrafts()..removeWhere((d) => d.id == id);
    await _prefs.setString(
        _kDraftsKey, json.encode(drafts.map((d) => d.toJson()).toList()));
  }
}

final draftStorageProvider = Provider<DraftStorage>((ref) {
  throw UnimplementedError(
      'Override draftStorageProvider in buildProviderOverrides');
});
