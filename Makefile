.DEFAULT_GOAL := help

help:
	@echo "TCP MESSAGE PROCESSOR - MAKEFILE"
	@echo ""
	@echo "DEVELOPMENT:"
	@echo " run-local       Start dependencies and run server locally"
	@echo " down            Stop all dependencies"
	@echo ""
	@echo "TESTING & QUALITY:"
	@echo " test            Run all tests"
	@echo " lint            Run golangci-lint"
	@echo " lint-fix        Run golangci-lint with auto-fix"
	@echo " mocks           Generate mocks with mockery"

run-local:
	@echo "Starting dependencies..."
	@docker-compose up -d postgres rabbitmq
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Starting server locally..."
	@go run cmd/server/main.go

down:
	@echo "Stopping dependencies..."
	@docker-compose down

test:
	@echo "Running all tests..."
	@go test -v -race ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@golangci-lint run --fix

mocks:
	@echo "Generating mocks with mockery..."
	@mockery
