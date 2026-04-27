# 10 · media-service

> Загрузка/выдача медиа (фото). Файлы лежат в S3-совместимом хранилище (MinIO в dev), метаданные — в Postgres.

## 1. Bounded Context

media-service владеет: media (метаданные: id, owner, kind, mime, size, width/height, key, status), upload_sessions (presigned).

Не владеет: где медиа используется (это знают places/journal/profile/trips через `*_media_id`).

## 2. Архитектура загрузки

Никаких multipart через сервис — Flutter загружает напрямую в S3 через presigned URL:

```
[Flutter] POST /v1/media/uploads { kind, mime, size }
       ──→ [media-service]
       creates row in `media_uploads` (status=pending), generates presigned PUT URL
       ←── { upload_id, upload_url, key, expires_at }

[Flutter] PUT <upload_url> Content-Type: <mime>  (raw bytes)

[Flutter] POST /v1/media/uploads/{id}/finalize { width, height, sha256 }
       ──→ [media-service]
       HEAD /<key> → проверяем размер
       создаём строку в `media`, status=ready, удаляем upload_session
       ←── { media_id }
```

После этого сервисы сохраняют `media_id` в свои таблицы.

## 3. Сущности

```go
type Media struct {
    ID         uuid.UUID
    OwnerID    uuid.UUID
    Kind       MediaKind  // photo | avatar (в MVP только эти)
    MIME       string     // image/jpeg | image/png | image/webp
    SizeBytes  int64
    Width      int
    Height     int
    Key        string     // s3 object key, формата "<owner_id>/<id>/<random>.<ext>"
    Status     string     // pending | ready | deleted
    Sha256     *string
    CreatedAt  time.Time
}

type UploadSession struct {
    ID         uuid.UUID
    OwnerID    uuid.UUID
    Kind       MediaKind
    MIME       string
    SizeBytes  int64
    Key        string     // pre-allocated
    ExpiresAt  time.Time
    Status     string     // pending | finalized | expired
    CreatedAt  time.Time
}
```

## 4. Use cases

- `RequestUpload(ctx, ownerID, kind, mime, sizeBytes)` →
  - Валидирует mime (whitelist `image/jpeg`, `image/png`, `image/webp`)
  - Валидирует size ≤ 15 MB
  - Генерирует key, presigned PUT URL TTL 10 мин
  - Сохраняет UploadSession
- `Finalize(ctx, ownerID, sessionID, width, height, sha256)` →
  - HEAD объекта в S3, проверяет размер
  - Создаёт Media, помечает session=finalized
  - Возвращает media_id
- `GetURL(ctx, viewerID, mediaID, transform)` →
  - `transform`: optional `?w=200&h=200&fit=cover` (через image proxy в Phase 3)
  - Для MVP: возвращаем presigned GET URL TTL 1 час прямо на оригинал
  - ACL: kind=avatar — открыто всем; kind=photo — нужно ownership-доказательство (см. ниже §6)
- `DeleteByOwner(ctx, ownerID, mediaID)` — soft delete + queue cleanup в S3.

## 5. HTTP API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| POST | `/v1/media/uploads` | user | Запрос presigned URL |
| POST | `/v1/media/uploads/{id}/finalize` | owner | Подтверждение загрузки |
| GET | `/v1/media/{id}` | viewer (см. ACL) | Получить presigned URL для просмотра |
| DELETE | `/v1/media/{id}` | owner | Удалить |
| GET | `/v1/media/internal/{id}/owner` | internal | Узнать owner_id (для других сервисов) |
| GET | `/v1/media/internal/by-ids?ids=...` | internal | Bulk-проверка существования и ownership |

## 6. ACL и приватность

- Аватары (`kind=avatar`) — публичные. URL может быть закэширован clientом без presigning (Phase 2: использовать публичный bucket policy для prefix `avatars/`).
- Фото записей (`kind=photo`) — приватные. `GET /v1/media/{id}` принимает запрос от **владельца** или от сервиса, доказавшего, что viewer имеет право (например, journal-service делает internal-запрос с `X-Viewer-Id` после проверки прав).
- В MVP: упрощаем — все photo URL доступны только владельцу. Шеринг (Phase 4) добавит token-based access.

## 7. БД-схема

```sql
SET search_path TO media, public;

CREATE TABLE media_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('photo','avatar')),
    mime TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','finalized','expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_media_uploads_owner ON media_uploads(owner_id) WHERE status='pending';

CREATE TABLE media_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    kind TEXT NOT NULL,
    mime TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    object_key TEXT NOT NULL UNIQUE,
    sha256 TEXT,
    status TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('ready','deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_media_owner ON media_media(owner_id, created_at DESC) WHERE status='ready';
```

## 8. Cron / Worker

- `expireUploads` каждые 5 минут: `UPDATE media_uploads SET status='expired' WHERE status='pending' AND expires_at < now()`.
- `cleanupS3Worker` (Phase 2): удаляет объекты с status=expired/deleted в S3 батчами.

## 9. S3/MinIO конфигурация

| Env | Default | Назначение |
|---|---|---|
| S3_ENDPOINT | `http://minio:9000` | для MinIO |
| S3_REGION | `us-east-1` | |
| S3_ACCESS_KEY | `minioadmin` | |
| S3_SECRET_KEY | `minioadmin` | |
| S3_BUCKET | `landmark-media` | |
| S3_USE_PATH_STYLE | `true` | для MinIO |
| MEDIA_MAX_SIZE_BYTES | `15728640` | 15 MB |
| MEDIA_PRESIGN_PUT_TTL | `10m` | |
| MEDIA_PRESIGN_GET_TTL | `1h` | |

Bucket создаётся автоматически в `docker-entrypoint.sh` через `mc mb` или вручную при старте `mc alias && mc mb`.

## 10. Чек-лист

- [ ] Запрос upload-URL валидирует mime/size, создаёт session
- [ ] Finalize создаёт Media и удаляет session
- [ ] presigned GET возвращает рабочий URL
- [ ] Cron expired works
- [ ] Storage отвязан абстракцией `Storage interface { PresignPut(key, mime, size, ttl) (url, err) ... }`, что позволит позже заменить MinIO на AWS S3 без правок в use cases

## 11. Связанные файлы

- profile (avatar) → `06`
- places (cover/photos) → `07`
- journal (entry photos) → `09`
- БД → `15`
