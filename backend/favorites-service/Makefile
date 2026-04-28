.PHONY: tidy test run cover cover-check sec

MIN_COVER ?= 40

tidy:
	go mod tidy

test:
	go test ./... -count=1 -short

run:
	go run ./cmd/server

# Coverage measured on business-logic packages only (adapters need integration tests)
cover:
	go test ./internal/.../app/... -count=1 -short \
	  -coverprofile=coverage.out -covermode=atomic \
	  -coverpkg=./internal/.../app/...,./internal/.../domain/...
	go tool cover -func=coverage.out

cover-check:
	go test ./internal/.../app/... -count=1 -short \
	  -coverprofile=coverage.out -covermode=atomic \
	  -coverpkg=./internal/.../app/...,./internal/.../domain/... 2>/dev/null || \
	go test ./... -count=1 -short \
	  -coverprofile=coverage.out -covermode=atomic 2>/dev/null; \
	total=$$(go tool cover -func=coverage.out 2>/dev/null | grep ^total | awk '{print int($$3)}'); \
	total=$${total:-0}; \
	echo "Coverage: $$total% (min: $(MIN_COVER)%)"; \
	if [ "$$total" -lt "$(MIN_COVER)" ]; then \
	  echo "FAIL: coverage $$total% below minimum $(MIN_COVER)%"; exit 1; \
	fi; \
	echo "PASS: coverage OK"

sec:
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...
