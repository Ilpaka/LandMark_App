FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src/backend
COPY places-service/go.mod places-service/go.sum ./places-service/
RUN cd places-service && go mod download
COPY places-service/ ./places-service/
WORKDIR /src/backend/places-service
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=build /out/server /out/migrate /app/
COPY places-service/db/migrations /app/db/migrations
COPY places-service/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh && chown -R app:app /app
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV GOOSE_MIGRATION_DIR=/app/db/migrations
ENTRYPOINT ["/app/docker-entrypoint.sh"]
