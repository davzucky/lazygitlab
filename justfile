app := "lazygitlab"

default:
    @just --list

build:
    go build -o {{app}} ./cmd/lazygitlab

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
