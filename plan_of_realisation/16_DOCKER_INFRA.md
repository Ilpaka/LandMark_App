# 16 · Docker-инфра, корневой compose, env, CI

> Один docker-compose в корне `backend/` поднимает всю систему. На dev-машине достаточно `cd backend && docker compose up -d`.

## 1. Структура

```
backend/
├── docker-compose.yml         ← корневой
├── .env.example               ← шаблон для разработчика
├── scripts/
│   ├── init-db.sql            ← создание users + schemas (см. 15)
│   └── init-minio.sh          ← создаёт bucket `landmark-media`
├── api-gateway/Dockerfile
├── auth-service/Dockerfile
├── ...
└── pkg/
    ├── observability/
    ├── httpx/
    ├── authmid/
    └── eventbus/
```

## 2. Корневой `docker-compose.yml`

```yaml
name: landmark

x-common-env: &common-env
  APP_ENV: development
  LOG_LEVEL: info
  REDIS_URL: redis://redis:6379/0
  OTEL_EXPORTER_OTLP_ENDPOINT: ${OTEL_EXPORTER_OTLP_ENDPOINT:-}
  INTERNAL_API_KEY: ${INTERNAL_API_KEY:-dev-internal-key-change-me}

x-go-build: &go-build
  restart: unless-stopped
  networks: [landmark-internal]

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: landmark
      POSTGRES_PASSWORD: landmark
      POSTGRES_DB: landmark
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/01-init.sql:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U landmark -d landmark"]
      interval: 3s
      timeout: 5s
      retries: 20
    networks: [landmark-internal]

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 3s
      timeout: 3s
      retries: 10
    networks: [landmark-internal]

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes: [miniodata:/data]
    ports:
      - "9001:9001"  # console
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/ready"]
      interval: 5s
      timeout: 5s
      retries: 10
    networks: [landmark-internal]

  minio-init:
    image: minio/mc:latest
    depends_on:
      minio: {condition: service_healthy}
    entrypoint: ["/bin/sh", "-c"]
    command:
      - |
        mc alias set local http://minio:9000 minioadmin minioadmin;
        mc mb -p local/landmark-media || true;
        mc anonymous set none local/landmark-media;
    networks: [landmark-internal]

  mailhog:
    image: mailhog/mailhog:latest
    ports:
      - "8025:8025"  # web UI
    networks: [landmark-internal]

  # ───── Backend services ───────────────────────────────────────────

  auth-service:
    <<: *go-build
    build:
      context: .
      dockerfile: auth-service/Dockerfile
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=auth,public
      JWT_PRIVATE_KEY_PATH: /app/keys/jwt_rsa.pem
      JWT_PUBLIC_KEY_PATH: /app/keys/jwt_rsa.pub.pem
      JWT_ISSUER: landmark.app
      JWT_AUDIENCE: landmark-clients
      AUTH_PEPPER: ${AUTH_PEPPER:-dev-pepper-change-me-32-bytes-min!}
      PROFILE_SERVICE_URL: http://profile-service:8080
      OTEL_SERVICE_NAME: landmark-auth-api
      SMTP_HOST: mailhog
      SMTP_PORT: "1025"
      SMTP_FROM: no-reply@landmark.local
    volumes:
      - authkeys:/app/keys
    depends_on:
      postgres: {condition: service_healthy}
      redis: {condition: service_healthy}

  profile-service:
    <<: *go-build
    build: { context: ., dockerfile: profile-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=profile,public
      OTEL_SERVICE_NAME: landmark-profile-api
    depends_on: { postgres: {condition: service_healthy} }

  places-service:
    <<: *go-build
    build: { context: ., dockerfile: places-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=places,public
      MEDIA_SERVICE_URL: http://media-service:8080
      OTEL_SERVICE_NAME: landmark-places-api
    depends_on: { postgres: {condition: service_healthy}, redis: {condition: service_healthy} }

  trips-service:
    <<: *go-build
    build: { context: ., dockerfile: trips-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=trips,public
      OTEL_SERVICE_NAME: landmark-trips-api
    depends_on: { postgres: {condition: service_healthy} }

  journal-service:
    <<: *go-build
    build: { context: ., dockerfile: journal-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=journal,public
      MEDIA_SERVICE_URL: http://media-service:8080
      TRIPS_SERVICE_URL: http://trips-service:8080
      OTEL_SERVICE_NAME: landmark-journal-api
    depends_on: { postgres: {condition: service_healthy} }

  media-service:
    <<: *go-build
    build: { context: ., dockerfile: media-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=media,public
      S3_ENDPOINT: http://minio:9000
      S3_REGION: us-east-1
      S3_ACCESS_KEY: minioadmin
      S3_SECRET_KEY: minioadmin
      S3_BUCKET: landmark-media
      S3_USE_PATH_STYLE: "true"
      OTEL_SERVICE_NAME: landmark-media-api
    depends_on:
      postgres: {condition: service_healthy}
      minio-init: {condition: service_completed_successfully}

  favorites-service:
    <<: *go-build
    build: { context: ., dockerfile: favorites-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=favorites,public
      OTEL_SERVICE_NAME: landmark-favorites-api
    depends_on: { postgres: {condition: service_healthy} }

  notifications-service:
    <<: *go-build
    build: { context: ., dockerfile: notifications-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=notifications,public
      SMTP_HOST: mailhog
      SMTP_PORT: "1025"
      SMTP_FROM: no-reply@landmark.local
      FCM_SERVICE_ACCOUNT_PATH: ${FCM_SERVICE_ACCOUNT_PATH:-}
      OTEL_SERVICE_NAME: landmark-notifications-api
    depends_on: { postgres: {condition: service_healthy}, redis: {condition: service_healthy} }

  moderation-service:
    <<: *go-build
    build: { context: ., dockerfile: moderation-service/Dockerfile }
    environment:
      <<: *common-env
      DATABASE_URL: postgres://landmark:landmark@postgres:5432/landmark?sslmode=disable&search_path=moderation,public
      PLACES_SERVICE_URL: http://places-service:8080
      OTEL_SERVICE_NAME: landmark-moderation-api
    depends_on: { postgres: {condition: service_healthy}, places-service: {condition: service_started} }

  api-gateway:
    <<: *go-build
    build: { context: ., dockerfile: api-gateway/Dockerfile }
    ports:
      - "8080:8080"
    environment:
      <<: *common-env
      AUTH_JWKS_URL: http://auth-service:8080/v1/auth/jwks
      JWT_ISSUER: landmark.app
      JWT_AUDIENCE: landmark-clients
      OTEL_SERVICE_NAME: landmark-api-gateway
      CORS_ORIGINS: ${CORS_ORIGINS:-*}
    depends_on:
      auth-service: {condition: service_started}

  # ───── Observability (профиль "obs") ─────────────────────────────

  otel-collector:
    profiles: [obs]
    image: otel/opentelemetry-collector-contrib:latest
    command: ["--config=/etc/otel-collector.yaml"]
    volumes:
      - ./scripts/otel-collector.yaml:/etc/otel-collector.yaml:ro
    networks: [landmark-internal]

  tempo:
    profiles: [obs]
    image: grafana/tempo:latest
    command: ["-config.file=/etc/tempo.yaml"]
    volumes:
      - ./scripts/tempo.yaml:/etc/tempo.yaml:ro
    networks: [landmark-internal]

  prometheus:
    profiles: [obs]
    image: prom/prometheus:latest
    volumes:
      - ./scripts/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    networks: [landmark-internal]

  grafana:
    profiles: [obs]
    image: grafana/grafana:latest
    ports: ["3000:3000"]
    networks: [landmark-internal]

volumes:
  pgdata:
  miniodata:
  authkeys:

networks:
  landmark-internal:
    driver: bridge
```

## 3. `.env.example`

```env
# Секреты и переключатели для dev. Скопируй в .env и переопредели.
AUTH_PEPPER=dev-pepper-change-me-32-bytes-min!
INTERNAL_API_KEY=dev-internal-key-change-me
CORS_ORIGINS=*
OTEL_EXPORTER_OTLP_ENDPOINT=
FCM_SERVICE_ACCOUNT_PATH=
```

## 4. Поднятие dev-окружения с нуля

```sh
cd /Users/ilpaka/Development/LandMark_App/backend
cp .env.example .env
docker compose up -d --build
docker compose logs -f auth-service api-gateway

# Проверка готовности:
curl http://localhost:8080/v1/auth/jwks
```

Опционально с observability:
```sh
docker compose --profile obs up -d
# Grafana: http://localhost:3000 (admin/admin), datasources prometheus + tempo
```

## 5. CI (GitHub Actions, пример)

`.github/workflows/backend-ci.yml`:

```yaml
name: backend-ci
on:
  push:
    paths: ['backend/**']
  pull_request:
    paths: ['backend/**']

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env: { POSTGRES_USER: landmark, POSTGRES_PASSWORD: landmark, POSTGRES_DB: landmark }
        ports: ["5432:5432"]
        options: >-
          --health-cmd "pg_isready -U landmark"
          --health-interval 5s --health-timeout 3s --health-retries 5
      redis:
        image: redis:7-alpine
        ports: ["6379:6379"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - name: Test all services
        run: |
          for s in backend/*/; do
            if [ -f "$s/go.mod" ]; then
              echo "::group::$s"
              (cd "$s" && go mod download && go test ./... -count=1 -short)
              echo "::endgroup::"
            fi
          done
      - name: Run integration (auth)
        env:
          DATABASE_URL: postgres://landmark:landmark@localhost:5432/landmark?sslmode=disable
          REDIS_URL: redis://localhost:6379/0
          AUTH_PEPPER: ci-pepper-12345678901234567890
        run: cd backend/auth-service && make test-integration
```

## 6. Health-проверки

- Каждый сервис имеет `GET /healthz` (унаследовано из Tennis: проверяет Postgres `SELECT 1`, Redis `PING`).
- В docker-compose добавляем `healthcheck:` (можно после стабилизации).

## 7. Связанные файлы

- DB-схемы → `15_DATABASE_SCHEMAS.md`
- Гейтвей → `14_API_GATEWAY.md`
- Сервисы → `05..13`
- Observability config → `22_OBSERVABILITY_SECURITY.md`
