# syntax=docker/dockerfile:1.7
FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct \
    GOSUMDB=sum.golang.google.cn
WORKDIR /src/backend
COPY notifications-service/go.mod notifications-service/go.sum ./notifications-service/
RUN --mount=type=cache,target=/go/pkg/mod,sharing=shared \
    cd notifications-service && go mod download
COPY notifications-service/ ./notifications-service/
WORKDIR /src/backend/notifications-service
RUN --mount=type=cache,target=/go/pkg/mod,sharing=shared \
    --mount=type=cache,target=/root/.cache/go-build,sharing=shared \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=build /out/server /out/migrate /app/
COPY notifications-service/db/migrations /app/db/migrations
COPY notifications-service/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh && chown -R app:app /app
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV GOOSE_MIGRATION_DIR=/app/db/migrations
ENTRYPOINT ["/app/docker-entrypoint.sh"]
