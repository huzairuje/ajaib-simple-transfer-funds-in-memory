.PHONY: help build run test test-unit test-integration test-e2e test-coverage clean fmt lint mocks help

help:
	@echo "Transfer Service - Makefile Commands"
	@echo "====================================="
	@echo ""
	@echo "Building:"
	@echo "  make build          Build the application"
	@echo "  make run            Run the application"
	@echo ""
	@echo "Testing:"
	@echo "  make test           Run all tests"
	@echo "  make test-unit      Run unit tests only"
	@echo "  make test-integration Run integration tests"
	@echo "  make test-e2e       Run end-to-end tests"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make test-coverage-html Generate HTML coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt            Format code"
	@echo "  make lint           Run linter"
	@echo "  make vet            Run go vet"
	@echo ""
	@echo "Mocks:"
	@echo "  make mocks          Generate all mocks"
	@echo "  make mocks-clean    Remove generated mocks"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean          Clean build artifacts"
	@echo "  make deps           Download dependencies"
	@echo "  make tidy           Tidy go modules"
	@echo ""

build:
	@echo "Building transfer service..."
	@go build -o bin/transfer-service cmd/gateway/main.go
	@echo "✓ Build complete: bin/transfer-service"

run:
	@echo "Starting transfer service..."
	@go run cmd/gateway/main.go

test: test-unit test-integration
	@echo "✓ All tests passed"

test-unit:
	@echo "Running unit tests..."
	@go test ./internal/adapters/core/... \
		./internal/adapters/app/... \
		./internal/adapters/framework/secondary/repository/... \
		-v
	@echo "✓ Unit tests passed"

test-integration:
	@echo "Running integration tests..."
	@go test ./test/integration/... -v
	@echo "✓ Integration tests passed"

test-e2e:
	@echo "Running end-to-end tests..."
	@go test ./test/e2e/... -v -parallel=1
	@echo "✓ E2E tests passed"

test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out
	@echo "✓ Coverage report generated"

test-coverage-html:
	@echo "Generating HTML coverage report..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...
	@echo "✓ Linting complete"

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet complete"

mocks:
	@echo "Generating mocks..."
	@which mockgen > /dev/null || (echo "Installing mockgen..." && go install github.com/golang/mock/mockgen@latest)
	@mockgen -source=internal/ports/core/transfer.go -destination=test/mocks/core/mock_transfer.go -package=mocks
	@mockgen -source=internal/ports/app/transfer.go -destination=test/mocks/app/mock_transfer.go -package=mocks
	@mockgen -source=internal/ports/secondary/db/transfer.go -destination=test/mocks/db/mock_transfer.go -package=mocks
	@mockgen -source=internal/ports/secondary/cache/idempotency.go -destination=test/mocks/cache/mock_idempotency.go -package=mocks
	@echo "✓ Mocks generated"

mocks-clean:
	@echo "Cleaning generated mocks..."
	@rm -rf test/mocks/
	@echo "✓ Mocks cleaned"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "✓ Dependencies downloaded"

tidy:
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "✓ Modules tidied"

docker-build:
	@echo "Building Docker image..."
	@docker build -t transfer-service:latest .
	@echo "✓ Docker image built"

docker-run:
	@echo "Running Docker container..."
	@docker run -p 3400:3400 transfer-service:latest

benchmark:
	@echo "Running benchmarks..."
	@go test ./... -bench=. -benchmem
	@echo "✓ Benchmarks complete"
