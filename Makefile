APP := lazygitlab

.PHONY: build run test lint fmt tidy tui-validate pre-commit hooks commit-check ci

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

pre-commit:
	./scripts/pre-commit-check.sh

hooks:
	./scripts/install-githooks.sh

commit-check:
	./scripts/check-conventional-commits.sh

ci: lint test build tui-validate
