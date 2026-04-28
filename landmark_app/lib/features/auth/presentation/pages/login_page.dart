import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/design/components/wl_button.dart';
import '../../../../core/design/components/wl_text_field.dart';
import '../providers/auth_provider.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  final _emailCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _obscure = true;
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _emailCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(authProvider.notifier).login(_emailCtrl.text.trim(), _passwordCtrl.text);
      // Router will redirect to /main/map via redirect
    } catch (e) {
      setState(() { _error = 'Неверный email или пароль'; });
    } finally {
      if (mounted) setState(() { _loading = false; });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Вход')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Text('Добро пожаловать', style: AppTypography.h2),
                const SizedBox(height: AppSpacing.xl),
                WlTextField(
                  label: 'Email',
                  controller: _emailCtrl,
                  keyboardType: TextInputType.emailAddress,
                  textInputAction: TextInputAction.next,
                  validator: (v) => (v?.isEmpty ?? true) ? 'Введите email' : null,
                ),
                const SizedBox(height: AppSpacing.md),
                WlTextField(
                  label: 'Пароль',
                  controller: _passwordCtrl,
                  obscureText: _obscure,
                  showToggle: true,
                  onToggleObscure: () => setState(() => _obscure = !_obscure),
                  textInputAction: TextInputAction.done,
                  validator: (v) => (v?.isEmpty ?? true) ? 'Введите пароль' : null,
                ),
                if (_error != null) ...[
                  const SizedBox(height: AppSpacing.md),
                  Text(_error!, style: AppTypography.bodySmall.copyWith(color: AppColors.error)),
                ],
                const SizedBox(height: AppSpacing.xl),
                WlButton(label: 'Войти', onPressed: _submit, loading: _loading),
                const SizedBox(height: AppSpacing.md),
                TextButton(
                  onPressed: () => context.go('/auth/register'),
                  child: const Text('Нет аккаунта? Зарегистрироваться'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
