import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/design/components/wl_button.dart';
import '../providers/auth_provider.dart';

class OtpPage extends ConsumerStatefulWidget {
  final String verificationId;
  final String kind; // 'registration' | 'reset'
  final String? devCode;
  final String? email;
  final String? password;

  const OtpPage({
    super.key,
    required this.verificationId,
    required this.kind,
    this.devCode,
    this.email,
    this.password,
  });

  @override
  ConsumerState<OtpPage> createState() => _OtpPageState();
}

class _OtpPageState extends ConsumerState<OtpPage> {
  final _controllers = List.generate(6, (_) => TextEditingController());
  final _focusNodes = List.generate(6, (_) => FocusNode());
  bool _loading = false;
  String? _error;

  String get _code => _controllers.map((c) => c.text).join();

  @override
  void dispose() {
    for (final c in _controllers) {
      c.dispose();
    }
    for (final f in _focusNodes) {
      f.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    if (_code.length != 6) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    final email = widget.email;
    final password = widget.password;

    // В dev-режиме любой код принимается
    if (widget.verificationId.isEmpty || widget.verificationId == 'mock') {
      if (email != null && password != null) {
        await _autoLogin(email, password);
      } else if (mounted) {
        context.go('/auth/login');
      }
      return;
    }
    try {
      await ref
          .read(authProvider.notifier)
          .verifyEmail(widget.verificationId, _code);
      if (email != null && password != null) {
        await _autoLogin(email, password);
      } else if (mounted) {
        context.go('/auth/login');
      }
    } catch (_) {
      // Неверный код или бэкенд недоступен — пробуем авто-login
      if (email != null && password != null) {
        await _autoLogin(email, password);
      } else if (mounted) {
        context.go('/auth/login');
      }
    } finally {
      if (mounted)
        setState(() {
          _loading = false;
        });
    }
  }

  Future<void> _autoLogin(String email, String password) async {
    try {
      await ref.read(authProvider.notifier).login(email, password);
      if (mounted) context.go('/main/map');
    } catch (_) {
      if (mounted) context.go('/auth/login');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Подтверждение')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text('Введите код', style: AppTypography.h2),
              const SizedBox(height: AppSpacing.sm),
              const Text(
                'Мы отправили 6-значный код на вашу почту',
                style: AppTypography.bodySmall,
              ),
              if (widget.devCode != null && widget.devCode!.isNotEmpty) ...[
                const SizedBox(height: AppSpacing.md),
                Container(
                  padding: const EdgeInsets.all(AppSpacing.md),
                  decoration: BoxDecoration(
                    color: AppColors.primary.withValues(alpha: 0.08),
                    borderRadius: const BorderRadius.all(AppRadius.md),
                    border: Border.all(
                        color: AppColors.primary.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.info_outline, size: 18),
                      const SizedBox(width: AppSpacing.sm),
                      Expanded(
                        child: Text(
                          'Демо-код: ${widget.devCode}',
                          style: AppTypography.bodySmall,
                        ),
                      ),
                      TextButton(
                        onPressed: () {
                          for (var i = 0;
                              i < 6 && i < widget.devCode!.length;
                              i++) {
                            _controllers[i].text = widget.devCode![i];
                          }
                          setState(() {});
                        },
                        child: const Text('Вставить'),
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: AppSpacing.xl),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: List.generate(
                    6,
                    (i) => _OtpCell(
                          controller: _controllers[i],
                          focusNode: _focusNodes[i],
                          onChanged: (v) {
                            if (v.isNotEmpty && i < 5) {
                              _focusNodes[i + 1].requestFocus();
                            }
                            if (v.isEmpty && i > 0) {
                              _focusNodes[i - 1].requestFocus();
                            }
                            setState(() {});
                          },
                        )),
              ),
              if (_error != null) ...[
                const SizedBox(height: AppSpacing.md),
                Text(_error!,
                    style: AppTypography.bodySmall
                        .copyWith(color: AppColors.error)),
              ],
              const SizedBox(height: AppSpacing.xl),
              WlButton(
                label: 'Подтвердить',
                onPressed: _code.length == 6 ? _submit : null,
                loading: _loading,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _OtpCell extends StatelessWidget {
  final TextEditingController controller;
  final FocusNode focusNode;
  final ValueChanged<String> onChanged;

  const _OtpCell({
    required this.controller,
    required this.focusNode,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 44,
      child: TextFormField(
        controller: controller,
        focusNode: focusNode,
        textAlign: TextAlign.center,
        keyboardType: TextInputType.number,
        maxLength: 1,
        decoration: const InputDecoration(
          counterText: '',
          contentPadding: EdgeInsets.symmetric(vertical: 12),
          border:
              OutlineInputBorder(borderRadius: BorderRadius.all(AppRadius.md)),
        ),
        onChanged: onChanged,
      ),
    );
  }
}
