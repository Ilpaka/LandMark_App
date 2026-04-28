.PHONY: tidy test run

tidy:
	go mod tidy

test:
	go test ./... -count=1 -short

run:
	go run ./cmd/server
