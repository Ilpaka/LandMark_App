import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/design/tokens.dart';
import '../../data/api/users_api.dart';
import '../providers/users_provider.dart';

/// Admin users management: list, search by email/phone, block/unblock,
/// force-logout, view audit log.
class UsersPage extends ConsumerStatefulWidget {
  const UsersPage({super.key});

  @override
  ConsumerState<UsersPage> createState() => _UsersPageState();
}

class _UsersPageState extends ConsumerState<UsersPage> {
  final _searchCtrl = TextEditingController();
  String? _query;

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  void _applyQuery() {
    final raw = _searchCtrl.text.trim();
    setState(() => _query = raw.isEmpty ? null : raw);
  }

  @override
  Widget build(BuildContext context) {
    final usersAsync = ref.watch(usersListProvider(_query));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Пользователи'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(usersListProvider(_query)),
          ),
        ],
      ),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: TextField(
              controller: _searchCtrl,
              textInputAction: TextInputAction.search,
              onSubmitted: (_) => _applyQuery(),
              decoration: InputDecoration(
                hintText: 'Поиск по email или телефону',
                prefixIcon: const Icon(Icons.search, size: 20),
                isDense: true,
                filled: true,
                fillColor: AppColors.surface,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(10),
                  borderSide: const BorderSide(color: AppColors.border),
                ),
                suffixIcon: _searchCtrl.text.isEmpty
                    ? null
                    : IconButton(
                        icon: const Icon(Icons.clear, size: 18),
                        onPressed: () {
                          _searchCtrl.clear();
                          _applyQuery();
                        },
                      ),
              ),
            ),
          ),
          Expanded(
            child: usersAsync.when(
              loading: () => const Center(
                child: CircularProgressIndicator(color: AppColors.primary),
              ),
              error: (e, _) => Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text('Ошибка: $e', textAlign: TextAlign.center),
                      const SizedBox(height: 12),
                      ElevatedButton(
                        onPressed: () =>
                            ref.invalidate(usersListProvider(_query)),
                        child: const Text('Повторить'),
                      ),
                    ],
                  ),
                ),
              ),
              data: (users) => users.isEmpty
                  ? const _EmptyState()
                  : RefreshIndicator(
                      color: AppColors.primary,
                      onRefresh: () async => ref
                          .read(usersListProvider(_query).notifier)
                          .refreshSelf(),
                      child: ListView.separated(
                        padding: const EdgeInsets.all(16),
                        itemCount: users.length,
                        separatorBuilder: (_, __) =>
                            const SizedBox(height: 8),
                        itemBuilder: (_, i) => _UserTile(
                          user: users[i],
                          onTap: () => _openUserSheet(context, users[i]),
                        ),
                      ),
                    ),
            ),
          ),
        ],
      ),
    );
  }

  void _openUserSheet(BuildContext context, AdminUser user) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (_) => _UserSheet(user: user, query: _query),
    );
  }
}

class _UserTile extends StatelessWidget {
  const _UserTile({required this.user, required this.onTap});

  final AdminUser user;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final color = user.isBlocked
        ? Colors.red
        : user.isAdmin
            ? AppColors.primary
            : AppColors.textPrimary;

    return Card(
      elevation: 0,
      color: AppColors.surface,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          child: Row(
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                alignment: Alignment.center,
                child: Icon(
                  user.isAdmin
                      ? Icons.shield_outlined
                      : user.isBlocked
                          ? Icons.block
                          : Icons.person_outline,
                  color: color,
                  size: 20,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      user.displayLabel,
                      style: const TextStyle(
                          fontWeight: FontWeight.w600, fontSize: 14),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 2),
                    Wrap(
                      spacing: 6,
                      runSpacing: 4,
                      children: [
                        _StatusChip(label: user.role, kind: _ChipKind.role),
                        _StatusChip(
                            label: user.status, kind: _ChipKind.status),
                      ],
                    ),
                  ],
                ),
              ),
              const Icon(Icons.chevron_right, color: AppColors.textSecondary),
            ],
          ),
        ),
      ),
    );
  }
}

enum _ChipKind { role, status }

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.label, required this.kind});
  final String label;
  final _ChipKind kind;

  @override
  Widget build(BuildContext context) {
    Color base;
    if (kind == _ChipKind.role) {
      base = label == 'admin' ? AppColors.primary : AppColors.textSecondary;
    } else {
      switch (label) {
        case 'blocked':
          base = Colors.red;
          break;
        case 'pending_verification':
          base = Colors.orange;
          break;
        case 'deleted':
          base = AppColors.textSecondary;
          break;
        default:
          base = Colors.green;
      }
    }
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: base.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          color: base,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.search_off, size: 64, color: AppColors.textSecondary),
          SizedBox(height: 12),
          Text('Никого не найдено',
              style: TextStyle(color: AppColors.textSecondary)),
        ],
      ),
    );
  }
}

class _UserSheet extends ConsumerStatefulWidget {
  const _UserSheet({required this.user, required this.query});
  final AdminUser user;
  final String? query;

  @override
  ConsumerState<_UserSheet> createState() => _UserSheetState();
}

class _UserSheetState extends ConsumerState<_UserSheet> {
  bool _busy = false;

  Future<void> _run(Future<void> Function() op, String okMsg) async {
    setState(() => _busy = true);
    try {
      await op();
      if (!mounted) return;
      Navigator.pop(context);
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(okMsg)));
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text('Ошибка: $e')));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _block() async {
    final reason = await _askReason(context);
    if (reason == null) return;
    await _run(
      () => ref
          .read(usersListProvider(widget.query).notifier)
          .block(widget.user.id, reason),
      'Пользователь заблокирован',
    );
  }

  Future<void> _unblock() => _run(
        () => ref
            .read(usersListProvider(widget.query).notifier)
            .unblock(widget.user.id),
        'Пользователь разблокирован',
      );

  Future<void> _forceLogout() => _run(
        () => ref
            .read(usersListProvider(widget.query).notifier)
            .forceLogout(widget.user.id),
        'Все сессии завершены',
      );

  void _showAudit() {
    Navigator.pop(context);
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (_) => _AuditSheet(userId: widget.user.id),
    );
  }

  @override
  Widget build(BuildContext context) {
    final u = widget.user;
    return SafeArea(
      child: Padding(
        padding: EdgeInsets.fromLTRB(
          16,
          12,
          16,
          16 + MediaQuery.of(context).viewInsets.bottom,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: AppColors.border,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 12),
            Text(u.displayLabel,
                style: const TextStyle(
                    fontSize: 18, fontWeight: FontWeight.w700)),
            const SizedBox(height: 4),
            SelectableText(
              'ID: ${u.id}',
              style: const TextStyle(
                fontSize: 11,
                color: AppColors.textSecondary,
                fontFamily: 'monospace',
              ),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 6,
              children: [
                _StatusChip(label: u.role, kind: _ChipKind.role),
                _StatusChip(label: u.status, kind: _ChipKind.status),
                if (u.emailVerifiedAt != null)
                  const _StatusChip(label: 'email verified', kind: _ChipKind.status),
                if (u.phoneVerifiedAt != null)
                  const _StatusChip(label: 'phone verified', kind: _ChipKind.status),
              ],
            ),
            const SizedBox(height: 16),
            _kv('Email', u.email ?? '—'),
            _kv('Телефон', u.phoneE164 ?? '—'),
            _kv('Создан', _fmtDate(u.createdAt)),
            _kv('Последний вход',
                u.lastLoginAt == null ? '—' : _fmtDate(u.lastLoginAt!)),
            if (u.isBlocked) ...[
              _kv('Заблокирован', _fmtDate(u.blockedAt!)),
              _kv('Причина', u.blockedReason ?? '—'),
            ],
            const SizedBox(height: 20),
            if (u.isBlocked)
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.green,
                  foregroundColor: Colors.white,
                  minimumSize: const Size.fromHeight(46),
                ),
                icon: const Icon(Icons.lock_open, size: 18),
                label: const Text('Разблокировать'),
                onPressed: _busy ? null : _unblock,
              )
            else
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.red,
                  foregroundColor: Colors.white,
                  minimumSize: const Size.fromHeight(46),
                ),
                icon: const Icon(Icons.block, size: 18),
                label: const Text('Заблокировать'),
                onPressed: _busy ? null : _block,
              ),
            const SizedBox(height: 8),
            OutlinedButton.icon(
              style: OutlinedButton.styleFrom(
                minimumSize: const Size.fromHeight(46),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10)),
              ),
              icon: const Icon(Icons.logout, size: 18),
              label: const Text('Завершить все сессии'),
              onPressed: _busy ? null : _forceLogout,
            ),
            const SizedBox(height: 8),
            OutlinedButton.icon(
              style: OutlinedButton.styleFrom(
                minimumSize: const Size.fromHeight(46),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10)),
              ),
              icon: const Icon(Icons.history, size: 18),
              label: const Text('Журнал событий'),
              onPressed: _busy ? null : _showAudit,
            ),
          ],
        ),
      ),
    );
  }

  Widget _kv(String k, String v) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 110,
            child: Text(k,
                style: const TextStyle(
                    fontSize: 12, color: AppColors.textSecondary)),
          ),
          Expanded(
            child: Text(v,
                style:
                    const TextStyle(fontSize: 13, fontWeight: FontWeight.w500)),
          ),
        ],
      ),
    );
  }
}

Future<String?> _askReason(BuildContext context) async {
  final ctrl = TextEditingController();
  return showDialog<String>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: const Text('Причина блокировки'),
      content: TextField(
        controller: ctrl,
        autofocus: true,
        maxLines: 3,
        decoration: InputDecoration(
          hintText: 'Например: спам, нарушение правил…',
          border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
        ),
      ),
      actions: [
        TextButton(
            onPressed: () => Navigator.pop(ctx), child: const Text('Отмена')),
        ElevatedButton(
          style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red, foregroundColor: Colors.white),
          onPressed: () {
            final txt = ctrl.text.trim();
            if (txt.isEmpty) return;
            Navigator.pop(ctx, txt);
          },
          child: const Text('Заблокировать'),
        ),
      ],
    ),
  );
}

String _fmtDate(DateTime dt) {
  String two(int x) => x.toString().padLeft(2, '0');
  return '${two(dt.day)}.${two(dt.month)}.${dt.year} ${two(dt.hour)}:${two(dt.minute)}';
}

class _AuditSheet extends ConsumerWidget {
  const _AuditSheet({required this.userId});
  final String userId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncEvents = ref.watch(auditProvider(userId));
    return SafeArea(
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxHeight: MediaQuery.of(context).size.height * 0.8,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 12),
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: AppColors.border,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(height: 12),
            const Text('Журнал событий',
                style:
                    TextStyle(fontSize: 17, fontWeight: FontWeight.w700)),
            const SizedBox(height: 12),
            Expanded(
              child: asyncEvents.when(
                loading: () => const Center(
                  child: CircularProgressIndicator(color: AppColors.primary),
                ),
                error: (e, _) => Center(child: Text('Ошибка: $e')),
                data: (events) => events.isEmpty
                    ? const Center(
                        child: Text('Событий нет',
                            style:
                                TextStyle(color: AppColors.textSecondary)))
                    : ListView.separated(
                        padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                        itemCount: events.length,
                        separatorBuilder: (_, __) =>
                            const Divider(height: 14),
                        itemBuilder: (_, i) {
                          final e = events[i];
                          return Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Expanded(
                                    child: Text(e.eventType,
                                        style: const TextStyle(
                                            fontWeight: FontWeight.w600,
                                            fontSize: 13)),
                                  ),
                                  Text(_fmtDate(e.createdAt),
                                      style: const TextStyle(
                                          fontSize: 11,
                                          color: AppColors.textSecondary)),
                                ],
                              ),
                              if (e.ip != null) ...[
                                const SizedBox(height: 2),
                                Text('IP: ${e.ip}',
                                    style: const TextStyle(
                                        fontSize: 11,
                                        color: AppColors.textSecondary)),
                              ],
                              if (e.meta != null && e.meta!.isNotEmpty) ...[
                                const SizedBox(height: 4),
                                Text(
                                  e.meta!.entries
                                      .map((kv) => '${kv.key}: ${kv.value}')
                                      .join('  ·  '),
                                  style: const TextStyle(
                                      fontSize: 11,
                                      color: AppColors.textSecondary),
                                ),
                              ],
                            ],
                          );
                        },
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
