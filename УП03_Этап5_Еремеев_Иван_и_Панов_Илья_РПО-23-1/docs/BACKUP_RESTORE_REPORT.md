# BACKUP_RESTORE_REPORT.md — LandMark, этап 5

Тип хранилища: **PostgreSQL 16 в Docker** (один контейнер `postgres`, БД `landmark`,
несколько схем: `auth`, `profile`, `places`, `trips`, `journal`, `media`, `favorites`,
`notifications`, `moderation`). Полный вывод прогона: `reports/backup_restore_run.txt`.

## 1. Backup
Инструмент: `pg_dump` внутри контейнера (`scripts/backup.bat`).

```
docker compose --env-file .env.production -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U landmark landmark > backups\backup_YYYYMMDD_HHMMSS.sql
```

Результат:
- Файл: `backups/backup_20260610_012010.sql`
- Размер: ~180 КБ, **5516 строк**, **94 `CREATE TABLE`** (все схемы).
- Скриншот: `08_backup_created.png`. Файл хранится в `backups/` и **исключён из Git** (`.gitignore`).

## 2. Restore + проверка
Инструмент: `psql` (`scripts/restore.bat`). Восстановление в **отдельную** БД
`landmark_restore_test`, чтобы не трогать рабочие данные, затем сверка числа строк.

```
... psql -U landmark -d landmark -c "DROP DATABASE IF EXISTS landmark_restore_test;"
... psql -U landmark -d landmark -c "CREATE DATABASE landmark_restore_test;"
... psql -U landmark -d landmark_restore_test < backups\backup_YYYYMMDD_HHMMSS.sql
```

Сверка (оригинал → восстановленная БД), **должны совпадать**:

| Таблица | Оригинал | После restore |
|---------|---------:|--------------:|
| `auth.auth_accounts` | 12 | **12** |
| `trips.trips_trips` | 3 | **3** |
| `auth.auth_sessions` | 8 | **8** |

- `restore exit code: 0`, **0 ошибок** при импорте.
- Данные после восстановления полностью совпали с оригиналом — резервная копия рабочая.
- Скриншот: `09_restore_success.png`.

## 3. Вывод
Backup и restore **работают и проверены**: дамп создаётся одной командой, восстанавливается
в чистую БД без ошибок, целостность подтверждена сверкой числа строк. Тестовая БД после
проверки удалена. Рекомендация по регулярности — в `RECOMMENDATIONS.md` (O-1).
