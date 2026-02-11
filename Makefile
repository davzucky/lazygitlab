APP := lazygitlab

.PHONY: build run test lint fmt tidy tui-validate

build:
	go build -o $(APP) ./cmd/lazygitlab

run:
	go run ./cmd/lazygitlab

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

tui-validate:
	./scripts/validate-tui-drift.sh
