.PHONY: default run build test
# Variables
APP_NAME=jorge-cli

# Tasks
default: run-api

run-api:
	@swag init -g ./cmd/api/main.go --parseDependency -parseInternal
	@go run ./cmd/api/main.go
build:
	@go build -o $(APP_NAME) main.go
test:
	@go test ./ ...