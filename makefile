.PHONY: default run build test test-unit test-integration test-e2e test-coverage clean

# Variables
APP_NAME=jorge-cli
CLI_PATH=./cmd/cli
API_PATH=./cmd/api

# Tasks
default: run-api

# Run API server
run-api:
	@swag init -g $(API_PATH)/main.go --parseDependency -parseInternal
	@go run $(API_PATH)/main.go

# Run CLI
run-cli:
	@go run $(CLI_PATH)/main.go

# Build
build:
	@go build -o $(APP_NAME) $(CLI_PATH)/main.go

build-api:
	@go build -o $(APP_NAME)-api $(API_PATH)/main.go

build-all: build build-api

# Testing
test: test-unit

test-unit:
	@go test -v -short ./...

test-race:
	@go test -v -race ./...

test-integration:
	@go test -v -tags=integration ./test/integration/...

test-e2e:
	@RUN_E2E_TESTS=true go test -v ./test/e2e/...

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Benchmarks
bench:
	@go test -bench=. -benchmem ./...

# Clean
clean:
	@rm -f $(APP_NAME) $(APP_NAME)-api coverage.out coverage.html
	@rm -rf ./db/*.db

# Development
fmt:
	@go fmt ./...

lint:
	@golangci-lint run ./...

# Help
help:
	@echo "Available targets:"
	@echo "  run-api         - Run API server with Swagger"
	@echo "  run-cli         - Run CLI"
	@echo "  build           - Build CLI binary"
	@echo "  build-api       - Build API binary"
	@echo "  build-all       - Build both CLI and API"
	@echo "  test            - Run unit tests"
	@echo "  test-unit       - Run unit tests"
	@echo "  test-race       - Run tests with race detector"
	@echo "  test-integration- Run integration tests"
	@echo "  test-e2e        - Run end-to-end tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  bench           - Run benchmarks"
	@echo "  clean           - Clean build artifacts"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter"
