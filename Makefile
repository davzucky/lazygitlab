.PHONY: build run test lint clean install help

BINARY_NAME=lazygitlab
BUILD_DIR=bin
CMD_DIR=cmd/lazygitlab
MAIN_FILE=$(CMD_DIR)/main.go
MAIN_PKG=./$(CMD_DIR)

help:
	@echo "Available targets:"
	@echo "  make build   - Build the binary"
	@echo "  make run     - Build and run the application"
	@echo "  make test    - Run tests"
	@echo "  make lint    - Run linters"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make install - Install the binary to GOPATH/bin"

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install $(MAIN_PKG)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download
