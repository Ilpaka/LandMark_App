import 'package:flutter/material.dart';
import '../../../../core/design/tokens.dart';

class HelpPage extends StatelessWidget {
  const HelpPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Помощь'),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        children: const [
          _Section(
            title: 'Карта',
            items: [
              _HelpItem(
                icon: Icons.map_outlined,
                question: 'Как найти место на карте?',
                answer:
                    'Нажмите на значок поиска в верхней части карты и введите название места, города или ключевое слово.',
              ),
              _HelpItem(
                icon: Icons.add_location_alt_outlined,
                question: 'Как предложить новое место?',
                answer:
                    'На экране карты нажмите кнопку «Предложить место». Заполните название, координаты и описание. После проверки модератором место появится на карте.',
              ),
            ],
          ),
          _Section(
            title: 'Поездки',
            items: [
              _HelpItem(
                icon: Icons.luggage_outlined,
                question: 'Как создать поездку?',
                answer:
                    'Перейдите на вкладку «Поездки» и нажмите +. Введите название, описание и даты. Затем добавляйте остановки с координатами.',
              ),
              _HelpItem(
                icon: Icons.auto_fix_high,
                question: 'Что такое оптимизация маршрута?',
                answer:
                    'Кнопка ✦ в карточке поездки автоматически перестраивает порядок остановок, чтобы минимизировать общую дистанцию.',
              ),
            ],
          ),
          _Section(
            title: 'Журнал',
            items: [
              _HelpItem(
                icon: Icons.book_outlined,
                question: 'Как добавить запись?',
                answer:
                    'На вкладке «Журнал» нажмите +. Можно добавить заголовок, текст, настроение и прикрепить фотографии.',
              ),
              _HelpItem(
                icon: Icons.wifi_off,
                question: 'Что происходит с записями без интернета?',
                answer:
                    'Записи сохраняются как черновики на устройстве. При восстановлении соединения они автоматически синхронизируются.',
              ),
            ],
          ),
          _Section(
            title: 'Профиль',
            items: [
              _HelpItem(
                icon: Icons.person_outline,
                question: 'Как изменить аватар?',
                answer:
                    'Перейдите в «Профиль» → «Редактировать». Нажмите на аватар и выберите фото из галереи.',
              ),
              _HelpItem(
                icon: Icons.notifications_outlined,
                question: 'Как настроить уведомления?',
                answer:
                    'В «Настройках» → «Уведомления» можно включить или отключить push-уведомления о поездках, модерации и маркетинговые email.',
              ),
              _HelpItem(
                icon: Icons.near_me_outlined,
                question: 'Что такое уведомления о близких местах?',
                answer:
                    'Когда вы окажетесь в радиусе 500 м от избранного POI, приложение отправит уведомление. Можно включить в Настройки → Геолокация.',
              ),
            ],
          ),
          _Section(
            title: 'Аккаунт',
            items: [
              _HelpItem(
                icon: Icons.lock_outline,
                question: 'Как изменить пароль?',
                answer:
                    'На экране входа нажмите «Забыли пароль?». Введите email — вам придёт код для сброса пароля.',
              ),
              _HelpItem(
                icon: Icons.delete_forever_outlined,
                question: 'Как удалить аккаунт?',
                answer:
                    'Профиль → прокрутите вниз → «Удалить аккаунт». Все данные будут безвозвратно удалены.',
              ),
            ],
          ),
          SizedBox(height: 24),
          _ContactTile(),
          SizedBox(height: 24),
        ],
      ),
    );
  }
}

class _Section extends StatelessWidget {
  final String title;
  final List<_HelpItem> items;
  const _Section({required this.title, required this.items});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(0, 20, 0, 8),
          child: Text(
            title,
            style: const TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.w600,
              color: AppColors.textSecondary,
              letterSpacing: 0.5,
            ),
          ),
        ),
        Card(
          margin: EdgeInsets.zero,
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          elevation: 0,
          color: AppColors.surface,
          child: Column(
            children: [
              for (var i = 0; i < items.length; i++) ...[
                if (i > 0) const Divider(height: 1, indent: 56),
                items[i],
              ],
            ],
          ),
        ),
      ],
    );
  }
}

class _HelpItem extends StatefulWidget {
  final IconData icon;
  final String question;
  final String answer;
  const _HelpItem(
      {required this.icon, required this.question, required this.answer});

  @override
  State<_HelpItem> createState() => _HelpItemState();
}

class _HelpItemState extends State<_HelpItem> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () => setState(() => _expanded = !_expanded),
      borderRadius: BorderRadius.circular(12),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(widget.icon, size: 22, color: AppColors.primary),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    widget.question,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ),
                AnimatedRotation(
                  turns: _expanded ? 0.5 : 0,
                  duration: const Duration(milliseconds: 200),
                  child: const Icon(Icons.keyboard_arrow_down,
                      color: AppColors.textSecondary),
                ),
              ],
            ),
            if (_expanded) ...[
              const SizedBox(height: 8),
              Padding(
                padding: const EdgeInsets.only(left: 34),
                child: Text(
                  widget.answer,
                  style: const TextStyle(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                    height: 1.5,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _ContactTile extends StatelessWidget {
  const _ContactTile();

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      elevation: 0,
      color: AppColors.primary.withValues(alpha: 0.08),
      child: const Padding(
        padding: EdgeInsets.all(16),
        child: Row(
          children: [
            Icon(Icons.mail_outline, color: AppColors.primary),
            SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Есть вопросы?',
                      style: TextStyle(
                          fontWeight: FontWeight.w600,
                          color: AppColors.primary)),
                  SizedBox(height: 2),
                  Text('support@wanderlog.app',
                      style: TextStyle(fontSize: 13, color: AppColors.primary)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
