import 'dart:async';
import 'dart:math';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:geolocator/geolocator.dart';

const double _kRadiusMeters = 300.0;
const int _kCooldownSeconds = 3600;

class WatchedPlace {
  final String id;
  final String title;
  final String? city;
  final double latitude;
  final double longitude;

  const WatchedPlace({
    required this.id,
    required this.title,
    this.city,
    required this.latitude,
    required this.longitude,
  });
}

class ProximityService {
  final FlutterLocalNotificationsPlugin _notifs =
      FlutterLocalNotificationsPlugin();

  StreamSubscription<Position>? _positionSub;
  final Map<String, DateTime> _notifiedAt = {};
  List<WatchedPlace> watchedPlaces = [];

  Future<void> init() async {
    const androidInit = AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosInit = DarwinInitializationSettings();
    await _notifs.initialize(
      const InitializationSettings(android: androidInit, iOS: iosInit),
    );
    await _notifs
        .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin>()
        ?.createNotificationChannel(const AndroidNotificationChannel(
          'proximity',
          'Рядом с местами',
          description: 'Уведомления при приближении к избранным местам',
          importance: Importance.defaultImportance,
        ));
  }

  Future<bool> requestPermission() async {
    final bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) return false;

    LocationPermission perm = await Geolocator.checkPermission();
    if (perm == LocationPermission.denied) {
      perm = await Geolocator.requestPermission();
    }
    return perm == LocationPermission.whileInUse ||
        perm == LocationPermission.always;
  }

  Future<bool> start(List<WatchedPlace> places) async {
    final granted = await requestPermission();
    if (!granted) return false;

    watchedPlaces = places;
    await _positionSub?.cancel();
    _positionSub = Geolocator.getPositionStream(
      locationSettings: const LocationSettings(
        accuracy: LocationAccuracy.medium,
        distanceFilter: 50,
      ),
    ).listen(_onPosition);
    return true;
  }

  Future<void> stop() async {
    await _positionSub?.cancel();
    _positionSub = null;
  }

  void _onPosition(Position pos) {
    for (final place in watchedPlaces) {
      final dist = _haversine(
          pos.latitude, pos.longitude, place.latitude, place.longitude);
      if (dist <= _kRadiusMeters) {
        _maybeNotify(place);
      }
    }
  }

  void _maybeNotify(WatchedPlace place) {
    final now = DateTime.now();
    final last = _notifiedAt[place.id];
    if (last != null && now.difference(last).inSeconds < _kCooldownSeconds) {
      return;
    }
    _notifiedAt[place.id] = now;
    _notifs.show(
      place.id.hashCode & 0x7fffffff,
      'Вы рядом: ${place.title}',
      place.city != null ? '📍 ${place.city}' : 'Посетите это место',
      const NotificationDetails(
        android: AndroidNotificationDetails(
          'proximity',
          'Рядом с местами',
          importance: Importance.defaultImportance,
          priority: Priority.defaultPriority,
        ),
        iOS: DarwinNotificationDetails(),
      ),
    );
  }

  double _haversine(double lat1, double lon1, double lat2, double lon2) {
    const r = 6371000.0;
    final dLat = _toRad(lat2 - lat1);
    final dLon = _toRad(lon2 - lon1);
    final a = sin(dLat / 2) * sin(dLat / 2) +
        cos(_toRad(lat1)) * cos(_toRad(lat2)) * sin(dLon / 2) * sin(dLon / 2);
    return r * 2 * atan2(sqrt(a), sqrt(1 - a));
  }

  double _toRad(double deg) => deg * pi / 180;
}
