# syntax=docker/dockerfile:1.7
# Build from backend/ directory:
#   docker compose build
FROM golang:1.25-alpine AS build
RUN apk add --no-cache git ca-certificates
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct \
    GOSUMDB=sum.golang.google.cn
WORKDIR /src/backend
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY pkg/observability ./pkg/observability/
WORKDIR /src/backend/auth-service
RUN --mount=type=cache,target=/go/pkg/mod,sharing=shared \
    go mod download
COPY auth-service/ .
RUN --mount=type=cache,target=/go/pkg/mod,sharing=shared \
    --mount=type=cache,target=/root/.cache/go-build,sharing=shared \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates openssl && \
	addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=build /out/server /out/migrate /app/
COPY auth-service/db/migrations /app/db/migrations
COPY auth-service/api/openapi /app/api/openapi
COPY auth-service/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh && \
	mkdir -p /app/keys && chown -R app:app /app
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080 \
	GOOSE_MIGRATION_DIR=/app/db/migrations \
	JWT_PRIVATE_KEY_PATH=/app/keys/jwt_rsa.pem \
	JWT_PUBLIC_KEY_PATH=/app/keys/jwt_rsa.pub.pem \
	OPENAPI_AUTH_YAML=/app/api/openapi/auth.yaml
ENTRYPOINT ["/app/docker-entrypoint.sh"]
