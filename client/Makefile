.PHONY: help run-local run-multi test test-integration test-all lint lint-fix clean stop-clients

.DEFAULT_GOAL := help

help:
	@echo "TCP MESSAGE PROCESSOR CLIENT - MAKEFILE"
	@echo ""
	@echo "DEVELOPMENT:"
	@echo " run-local        Run a single client"
	@echo " run-multi        Launch multiple clients"
	@echo " stop-clients     Kill all running client processes"
	@echo ""
	@echo "TESTING:"
	@echo " test             Run unit tests"
	@echo " test-integration Run integration tests"
	@echo " test-all         Run all tests (unit + integration)"
	@echo ""
	@echo "QUALITY:"
	@echo " lint             Run golangci-lint"
	@echo ""
	@echo "UTILS:"
	@echo " clean            Clean build artifacts"

run-local: stop-clients
	@echo "Starting client..."
	@go run cmd/client/main.go

run-multi: stop-clients
	@echo "Starting ${CLIENT_COUNT:-3} clients..."
	@./scripts/run-clients.sh

test:
	@echo "Running unit tests..."
	@go test ./internal/... -v

test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration ./test/integration/... -v -timeout 3m

test-all:
	@echo "Running all tests..."
	@$(MAKE) test
	@$(MAKE) test-integration

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin logs/client*.log
	@echo "Clean complete"

stop-clients:
	@echo "Stopping all running client processes..."
	@pkill -f "go run cmd/client/main.go" || true
	@echo "Clients stopped"

