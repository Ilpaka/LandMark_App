FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src/backend
COPY profile-service/go.mod profile-service/go.sum ./profile-service/
RUN cd profile-service && go mod download
COPY profile-service/ ./profile-service/
WORKDIR /src/backend/profile-service
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=build /out/server /out/migrate /app/
COPY profile-service/db/migrations /app/db/migrations
COPY profile-service/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh && chown -R app:app /app
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV GOOSE_MIGRATION_DIR=/app/db/migrations
ENTRYPOINT ["/app/docker-entrypoint.sh"]
