.PHONY: default run build test
# Variables
APP_NAME=jorge-cli

# Tasks
default: run

run:
	@swag init
	@go run main.go
build:
	@go build -o $(APP_NAME) main.go
test:
	@go test ./ ...