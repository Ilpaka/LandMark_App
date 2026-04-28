FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src/backend
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
RUN cd api-gateway && go mod download
COPY api-gateway/ ./api-gateway/
WORKDIR /src/backend/api-gateway
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=build /out/server /app/server
COPY api-gateway/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh && chown -R app:app /app
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080
ENTRYPOINT ["/app/docker-entrypoint.sh"]
