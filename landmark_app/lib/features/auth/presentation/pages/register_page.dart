import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/design/components/wl_button.dart';
import '../../../../core/design/components/wl_text_field.dart';
import '../../../../core/services/otp_notifier.dart';
import '../providers/auth_provider.dart';

class RegisterPage extends ConsumerStatefulWidget {
  const RegisterPage({super.key});

  @override
  ConsumerState<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends ConsumerState<RegisterPage> {
  final _formKey = GlobalKey<FormState>();
  final _nameCtrl = TextEditingController();
  final _emailCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _obscure = true;
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _nameCtrl.dispose();
    _emailCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final result = await ref.read(authProvider.notifier).register(
            _nameCtrl.text.trim(),
            _emailCtrl.text.trim(),
            _passwordCtrl.text,
          );
      final devCode = result.devCode;
      if (devCode != null && devCode.isNotEmpty) {
        unawaited(OtpNotifier.showOtp(devCode));
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Демо-код подтверждения: $devCode'),
              duration: const Duration(seconds: 30),
              behavior: SnackBarBehavior.floating,
            ),
          );
        }
      }
      if (mounted) {
        context.go('/auth/otp', extra: {
          'verification_id': result.verificationId,
          'kind': 'registration',
          'dev_code': devCode,
          'email': _emailCtrl.text.trim(),
          'password': _passwordCtrl.text,
        });
      }
    } catch (_) {
      setState(() {
        _error = 'Ошибка регистрации. Проверьте данные и попробуйте снова.';
      });
    } finally {
      if (mounted)
        setState(() {
          _loading = false;
        });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Регистрация')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Text('Создать аккаунт', style: AppTypography.h2),
                const SizedBox(height: AppSpacing.xl),
                WlTextField(
                  label: 'Имя',
                  controller: _nameCtrl,
                  textInputAction: TextInputAction.next,
                  validator: (v) => (v?.isEmpty ?? true) ? 'Введите имя' : null,
                ),
                const SizedBox(height: AppSpacing.md),
                WlTextField(
                  label: 'Email',
                  controller: _emailCtrl,
                  keyboardType: TextInputType.emailAddress,
                  textInputAction: TextInputAction.next,
                  validator: (v) {
                    if (v == null || v.isEmpty) return 'Введите email';
                    if (!v.contains('@')) return 'Некорректный email';
                    return null;
                  },
                ),
                const SizedBox(height: AppSpacing.md),
                WlTextField(
                  label: 'Пароль',
                  controller: _passwordCtrl,
                  obscureText: _obscure,
                  showToggle: true,
                  onToggleObscure: () => setState(() => _obscure = !_obscure),
                  textInputAction: TextInputAction.done,
                  validator: (v) {
                    if (v == null || v.length < 10)
                      return 'Минимум 10 символов';
                    if (!RegExp(r'[A-Za-z]').hasMatch(v)) return 'Нужна буква';
                    if (!RegExp(r'[0-9]').hasMatch(v)) return 'Нужна цифра';
                    return null;
                  },
                ),
                if (_error != null) ...[
                  const SizedBox(height: AppSpacing.md),
                  Text(_error!,
                      style: AppTypography.bodySmall
                          .copyWith(color: AppColors.error)),
                ],
                const SizedBox(height: AppSpacing.xl),
                WlButton(
                    label: 'Зарегистрироваться',
                    onPressed: _submit,
                    loading: _loading),
                const SizedBox(height: AppSpacing.md),
                TextButton(
                  onPressed: () => context.go('/auth/login'),
                  child: const Text('Уже есть аккаунт? Войти'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
