.PHONY: tidy test test-cover cover-check test-integration test-e2e lint sec migrate-up migrate-down run keys docker-up

tidy:
	go mod tidy

keys:
	mkdir -p keys && openssl genrsa -out keys/jwt_rsa.pem 2048 && openssl rsa -in keys/jwt_rsa.pem -pubout -out keys/jwt_rsa.pub.pem

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

test:
	go test ./... -count=1 -short

test-cover:
	go test ./... -count=1 -short -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 5

# Fails if total statement coverage is below MIN_COVER (default 12).
cover-check:
	go test ./... -count=1 -short -coverprofile=coverage.out
	@bash -c 'pct=$$(go tool cover -func=coverage.out | tail -1 | grep -oE "[0-9]+(\\.[0-9]+)?%" | tail -1 | tr -d "%"); \
	min=$${MIN_COVER:-12}; \
	if awk -v p="$$pct" -v m="$$min" "BEGIN{exit !(p+0 < m+0)}"; then echo "coverage $$pct% < $$min%"; exit 1; fi; \
	echo "coverage OK: $$pct%"'

test-integration:
	go test ./tests/integration/... -count=1 -tags=integration -timeout=400s

test-e2e:
	go test ./tests/e2e/... -count=1 -tags=e2e -timeout=600s -parallel 1

lint:
	go vet ./...

sec:
	go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet -exclude-generated -severity=high ./...

run:
	go run ./cmd/server

# From backend/auth-service/: Postgres (host :55432), Redis (:56379), Mailhog, auth-api (:8080). Override: TENNIS_POSTGRES_PORT, etc.
docker-up:
	docker compose -f ../../docker-compose.yml up -d --build
