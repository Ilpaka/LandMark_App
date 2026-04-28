import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../../../../core/design/tokens.dart';
import '../../../../core/networking/api_client.dart';
import '../../../journal/data/api/media_api.dart';
import '../../domain/entities/profile.dart';
import '../providers/profile_provider.dart';

class EditProfilePage extends ConsumerStatefulWidget {
  final UserProfile? profile;
  const EditProfilePage({super.key, this.profile});

  @override
  ConsumerState<EditProfilePage> createState() => _EditProfilePageState();
}

class _EditProfilePageState extends ConsumerState<EditProfilePage> {
  final _formKey = GlobalKey<FormState>();
  late final _nicknameCtrl = TextEditingController(text: widget.profile?.nickname ?? '');
  late final _displayNameCtrl = TextEditingController(text: widget.profile?.displayName ?? '');
  late final _bioCtrl = TextEditingController(text: widget.profile?.bio ?? '');
  late final _cityCtrl = TextEditingController(text: widget.profile?.city ?? '');
  late final _countryCtrl = TextEditingController(text: widget.profile?.country ?? '');
  bool _loading = false;
  File? _avatarFile;
  String? _newAvatarId;

  @override
  void dispose() {
    _nicknameCtrl.dispose();
    _displayNameCtrl.dispose();
    _bioCtrl.dispose();
    _cityCtrl.dispose();
    _countryCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickAvatar() async {
    final picker = ImagePicker();
    final picked = await picker.pickImage(source: ImageSource.gallery, maxWidth: 512, imageQuality: 85);
    if (picked == null) return;
    setState(() => _avatarFile = File(picked.path));
    try {
      final mediaApi = MediaApi(ref.read(apiClientProvider).dio);
      final id = await mediaApi.uploadFile(_avatarFile!, purpose: 'avatar');
      setState(() => _newAvatarId = id);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка загрузки аватара: $e')),
        );
      }
      setState(() => _avatarFile = null);
    }
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _loading = true);
    try {
      await ref.read(myProfileProvider.notifier).save(
        nickname: _nicknameCtrl.text.trim(),
        displayName: _displayNameCtrl.text.trim(),
        bio: _bioCtrl.text.trim().isEmpty ? null : _bioCtrl.text.trim(),
        city: _cityCtrl.text.trim().isEmpty ? null : _cityCtrl.text.trim(),
        country: _countryCtrl.text.trim().isEmpty ? null : _countryCtrl.text.trim(),
        avatarMediaId: _newAvatarId,
      );
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Редактировать профиль'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          TextButton(
            onPressed: _loading ? null : _save,
            child: const Text('Сохранить', style: TextStyle(color: AppColors.primary, fontWeight: FontWeight.w600)),
          ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Center(
              child: GestureDetector(
                onTap: _pickAvatar,
                child: Stack(
                  children: [
                    CircleAvatar(
                      radius: 48,
                      backgroundColor: AppColors.primary.withValues(alpha: 0.15),
                      backgroundImage: _avatarFile != null
                          ? FileImage(_avatarFile!) as ImageProvider
                          : null,
                      child: _avatarFile == null
                          ? Text(
                              (widget.profile?.displayName.isNotEmpty == true
                                  ? widget.profile!.displayName[0]
                                  : '?').toUpperCase(),
                              style: const TextStyle(
                                fontSize: 36,
                                color: AppColors.primary,
                                fontWeight: FontWeight.w600,
                              ),
                            )
                          : null,
                    ),
                    Positioned(
                      bottom: 0,
                      right: 0,
                      child: Container(
                        padding: const EdgeInsets.all(6),
                        decoration: const BoxDecoration(
                          color: AppColors.primary,
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(Icons.camera_alt, size: 16, color: Colors.white),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 20),
            _field(_nicknameCtrl, 'Никнейм', required: true,
                hint: 'Только буквы, цифры, _'),
            const SizedBox(height: 12),
            _field(_displayNameCtrl, 'Отображаемое имя', required: true),
            const SizedBox(height: 12),
            _field(_bioCtrl, 'О себе', maxLines: 3),
            const SizedBox(height: 12),
            _field(_cityCtrl, 'Город'),
            const SizedBox(height: 12),
            _field(_countryCtrl, 'Страна'),
            if (_loading) ...[
              const SizedBox(height: 24),
              const Center(child: CircularProgressIndicator(color: AppColors.primary)),
            ],
          ],
        ),
      ),
    );
  }

  Widget _field(TextEditingController ctrl, String label, {
    bool required = false,
    int maxLines = 1,
    String? hint,
  }) {
    return TextFormField(
      controller: ctrl,
      maxLines: maxLines,
      decoration: InputDecoration(
        labelText: label,
        hintText: hint,
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
        filled: true,
        fillColor: AppColors.surface,
      ),
      validator: required
          ? (v) => (v == null || v.trim().isEmpty) ? 'Обязательное поле' : null
          : null,
    );
  }
}
