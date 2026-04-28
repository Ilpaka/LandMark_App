import 'package:flutter/material.dart';
import '../tokens.dart';

class WlButton extends StatelessWidget {
  final String label;
  final VoidCallback? onPressed;
  final bool loading;
  final bool outlined;

  const WlButton({
    super.key,
    required this.label,
    this.onPressed,
    this.loading = false,
    this.outlined = false,
  });

  @override
  Widget build(BuildContext context) {
    if (outlined) {
      return OutlinedButton(
        onPressed: loading ? null : onPressed,
        style: OutlinedButton.styleFrom(
          foregroundColor: AppColors.primary,
          side: const BorderSide(color: AppColors.primary),
          minimumSize: const Size(double.infinity, 52),
          shape: const RoundedRectangleBorder(borderRadius: BorderRadius.all(AppRadius.md)),
        ),
        child: loading ? const _LoadingIndicator() : Text(label),
      );
    }
    return ElevatedButton(
      onPressed: loading ? null : onPressed,
      child: loading ? const _LoadingIndicator() : Text(label),
    );
  }
}

class _LoadingIndicator extends StatelessWidget {
  const _LoadingIndicator();
  @override
  Widget build(BuildContext context) =>
      const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2));
}
