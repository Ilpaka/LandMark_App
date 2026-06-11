import 'package:flutter/material.dart';

import '../design/tokens.dart';

/// Единое состояние ошибки списка/экрана: иконка, человеческое сообщение
/// и кнопка «Повторить». Не показываем пользователю сырой текст исключения.
class WlErrorState extends StatelessWidget {
  const WlErrorState({
    super.key,
    this.message = 'Не удалось загрузить данные',
    required this.onRetry,
  });

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.cloud_off, size: 48, color: AppColors.textSecondary),
          const SizedBox(height: 12),
          Text(
            message,
            textAlign: TextAlign.center,
            style: const TextStyle(color: AppColors.textSecondary),
          ),
          const SizedBox(height: 12),
          TextButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.refresh, size: 18),
            label: const Text('Повторить'),
          ),
        ],
      ),
    );
  }
}
