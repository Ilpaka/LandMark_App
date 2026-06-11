import 'package:flutter/material.dart';

import '../design/tokens.dart';

/// Кнопки управления картой: зум и (опционально) «моё местоположение».
/// Работают на любой платформе — в том числе там, где жесты pinch-zoom
/// недоступны (десктоп, эмулятор).
class MapZoomControls extends StatelessWidget {
  const MapZoomControls({
    super.key,
    required this.onZoomIn,
    required this.onZoomOut,
    this.onLocate,
    this.locating = false,
  });

  final VoidCallback onZoomIn;
  final VoidCallback onZoomOut;
  final VoidCallback? onLocate;
  final bool locating;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        _RoundButton(
          icon: Icons.add,
          tooltip: 'Приблизить',
          onTap: onZoomIn,
        ),
        const SizedBox(height: 8),
        _RoundButton(
          icon: Icons.remove,
          tooltip: 'Отдалить',
          onTap: onZoomOut,
        ),
        if (onLocate != null) ...[
          const SizedBox(height: 16),
          _RoundButton(
            icon: Icons.my_location,
            tooltip: 'Моё местоположение',
            onTap: locating ? null : onLocate,
            child: locating
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: AppColors.primary,
                    ),
                  )
                : null,
          ),
        ],
      ],
    );
  }
}

class _RoundButton extends StatelessWidget {
  const _RoundButton({
    required this.icon,
    required this.tooltip,
    required this.onTap,
    this.child,
  });

  final IconData icon;
  final String tooltip;
  final VoidCallback? onTap;
  final Widget? child;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: tooltip,
      child: Material(
        color: Colors.white,
        elevation: 3,
        shape: const CircleBorder(),
        child: InkWell(
          customBorder: const CircleBorder(),
          onTap: onTap,
          child: SizedBox(
            width: 44,
            height: 44,
            child: Center(
              child: child ?? Icon(icon, color: AppColors.primary, size: 22),
            ),
          ),
        ),
      ),
    );
  }
}
