import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'app.dart';
import 'bootstrap/di.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final overrides = await buildProviderOverrides();
  runApp(ProviderScope(overrides: overrides, child: const App()));
}
