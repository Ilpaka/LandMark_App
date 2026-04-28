import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/trip.dart';
import '../../data/api/trips_api.dart';
import '../../../../core/networking/api_client.dart';

final tripsApiProvider = Provider((ref) => TripsApi(ref.read(apiClientProvider).dio));

final tripsListProvider = FutureProvider.autoDispose<List<Trip>>((ref) {
  return ref.read(tripsApiProvider).listTrips();
});

final publicTripsProvider = FutureProvider.autoDispose<List<Trip>>((ref) {
  return ref.read(tripsApiProvider).listPublicTrips();
});

final tripRouteProvider = FutureProvider.autoDispose.family<List<Map<String, double>>, String>((ref, tripId) {
  return ref.read(tripsApiProvider).getRoute(tripId);
});
