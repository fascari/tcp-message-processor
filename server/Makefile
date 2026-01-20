.DEFAULT_GOAL := help

.PHONY: help run-local down test test-integration lint clean stop-servers

help:
	@echo "TCP MESSAGE PROCESSOR - MAKEFILE"
	@echo ""
	@echo "DEVELOPMENT:"
	@echo " run-local        Start dependencies and run server locally"
	@echo " down             Stop all dependencies"
	@echo " stop-servers     Kill all running server processes"
	@echo ""
	@echo "TESTING & QUALITY:"
	@echo " test             Run all tests"
	@echo " test-integration Run integration tests"
	@echo " lint             Run golangci-lint"
	@echo ""
	@echo "UTILS:"
	@echo " clean            Remove build artifacts and logs"

run-local: stop-servers
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

test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -v ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin
	@echo "Clean complete"

stop-servers:
	@echo "Stopping all running server processes..."
	@-lsof -ti:8888 | xargs kill -9 2>/dev/null || true
	@-pkill -f "go run cmd/server/main.go" 2>/dev/null || true
	@echo "Servers stopped"

