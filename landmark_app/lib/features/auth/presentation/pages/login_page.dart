import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/design/typography.dart';
import '../../../../core/design/components/wl_button.dart';
import '../../../../core/design/components/wl_text_field.dart';
import '../../../../core/networking/api_client.dart';
import '../../../../core/storage/secure_storage.dart';
import '../providers/auth_provider.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  static const _demoEmail = 'demo@apple.com';
  static const _demoPassword = 'DemoPass123456';

  final _formKey = GlobalKey<FormState>();
  final _emailCtrl = TextEditingController(text: _demoEmail);
  final _passwordCtrl = TextEditingController(text: _demoPassword);
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
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref
          .read(authProvider.notifier)
          .login(_emailCtrl.text.trim(), _passwordCtrl.text);
    } catch (e) {
      String msg;
      if (e is DioException) {
        final code = e.response?.statusCode;
        final body = e.response?.data?.toString() ?? e.message ?? '';
        msg = code != null ? 'HTTP $code · $body' : 'Сеть: ${e.message}';
      } else {
        msg = e.toString();
      }
      setState(() {
        _error = msg;
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
                  validator: (v) =>
                      (v?.isEmpty ?? true) ? 'Введите email' : null,
                ),
                const SizedBox(height: AppSpacing.md),
                WlTextField(
                  label: 'Пароль',
                  controller: _passwordCtrl,
                  obscureText: _obscure,
                  showToggle: true,
                  onToggleObscure: () => setState(() => _obscure = !_obscure),
                  textInputAction: TextInputAction.done,
                  validator: (v) =>
                      (v?.isEmpty ?? true) ? 'Введите пароль' : null,
                ),
                if (_error != null) ...[
                  const SizedBox(height: AppSpacing.md),
                  Text(_error!,
                      style: AppTypography.bodySmall
                          .copyWith(color: AppColors.error)),
                ],
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
                      const Expanded(
                        child: Text(
                          'Демо-аккаунт: demo@apple.com / DemoPass123456',
                          style: AppTypography.bodySmall,
                        ),
                      ),
                      TextButton(
                        onPressed: () {
                          _emailCtrl.text = _demoEmail;
                          _passwordCtrl.text = _demoPassword;
                          setState(() {});
                        },
                        child: const Text('Вставить'),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.xl),
                WlButton(label: 'Войти', onPressed: _submit, loading: _loading),
                const SizedBox(height: AppSpacing.md),
                TextButton(
                  onPressed: () => context.go('/auth/register'),
                  child: const Text('Нет аккаунта? Зарегистрироваться'),
                ),
                TextButton(
                  onPressed: () async {
                    final storage = ref.read(secureStorageProvider);
                    await storage.deleteAll();
                    final base = ref.read(apiClientProvider).dio.options.baseUrl;
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(SnackBar(
                        content: Text('Сессия очищена. API: $base'),
                      ));
                    }
                  },
                  child: const Text('Сбросить сессию (для отладки)'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
