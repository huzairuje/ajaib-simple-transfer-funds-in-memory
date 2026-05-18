# Transfer Service

A Go-based transfer service with idempotent status updates following Clean Architecture principles.

## Features

- Create transfers between accounts
- Get transfer details by ID
- List all transfers
- Update transfer status with idempotency support
- Clean Architecture with dependency inversion
- Thread-safe in-memory storage with mutex locks
- RESTful API using Gin framework
- Fully testable with mocked dependencies

## Architecture

This project follows **Clean Architecture** principles based on hexagonal architecture patterns.

```
cmd/gateway/                         # Application entry point
internal/
├── adapters/
│   ├── app/transfer/               # Application layer (use cases)
│   ├── core/                       # Core layer
│   │   ├── entity/                 # Domain entities
│   │   └── transfer/               # Business logic
│   └── framework/
│       ├── primary/rest_fiber/     # Primary adapters (HTTP handlers)
│       └── secondary/repository/   # Secondary adapters (DB, cache)
└── ports/                          # Interfaces/contracts
    ├── app/                        # Application ports
    ├── core/                       # Core ports
    └── secondary/                  # Secondary ports (DB, cache)
config/                             # Configuration
router/                             # HTTP routing
docs/                               # Documentation
```

## Getting Started

### Prerequisites

- Go 1.25.5 or higher
- Make (for using Makefile commands)
- Docker (optional, for containerization)
- Docker Compose (optional, for local development with Redis/PostgreSQL)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd ajaib-testing-code

# Install dependencies
make deps

# Build the application
make build
```

### Running the Application

```bash
# Using Make
make run

# Or directly with Go
go run cmd/gateway/main.go

# Using Docker
make docker-build
make docker-run

# Using Docker Compose (with Redis & PostgreSQL)
docker-compose up
```

The server will start on `http://localhost:3400`

## Makefile Commands

The project includes a comprehensive Makefile for common tasks:

```bash
# Show all available commands
make help

# Building
make build              # Build the application
make run                # Run the application

# Testing
make test               # Run all tests
make test-unit          # Run unit tests only
make test-integration   # Run integration tests
make test-e2e           # Run end-to-end tests
make test-coverage      # Run tests with coverage
make test-coverage-html # Generate HTML coverage report

# Code Quality
make fmt                # Format code
make lint               # Run linter
make vet                # Run go vet

# Mocks
make mocks              # Generate all mocks
make mocks-clean        # Remove generated mocks

# Maintenance
make clean              # Clean build artifacts
make deps               # Download dependencies
make tidy               # Tidy go modules
```

## API Endpoints

### Create Transfer

```bash
POST /v1/transfers
Content-Type: application/json

{
  "from": 1001,
  "to": 1002,
  "amount": 50000,
  "currency": "IDR",
  "from_balance": 100000,
  "to_balance": 50000
}
```

**Response:**
```json
{
  "transfer_id": 1,
  "status": "success",
  "from_balance": 50000,
  "to_balance": 100000
}
```

### Get Transfer by ID

```bash
GET /v1/transfers/:id
```

**Response:**
```json
{
  "id": 1,
  "from_account_id": 1001,
  "to_account_id": 1002,
  "amount": 50000,
  "currency": "IDR",
  "status": "success",
  "from_balance": 50000,
  "to_balance": 100000
}
```

### List All Transfers

```bash
GET /v1/transfers
```

**Response:**
```json
[
  {
    "id": 1,
    "from_account_id": 1001,
    "to_account_id": 1002,
    "amount": 50000,
    "currency": "IDR",
    "status": "success",
    "from_balance": 50000,
    "to_balance": 100000
  }
]
```

### Update Transfer Status (Idempotent)

```bash
PATCH /v1/transfers/:id/status
Content-Type: application/json

{
  "status": "completed"
}
```

**Response:**
```json
{
  "id": 1,
  "from_account_id": 1001,
  "to_account_id": 1002,
  "amount": 50000,
  "currency": "IDR",
  "status": "completed",
  "from_balance": 50000,
  "to_balance": 100000
}
```

**Idempotency:** Duplicate requests with the same transfer ID and status return the same result without re-processing.

## Idempotency

The service implements idempotency for status updates using internally generated keys:

- **Key Format:** `transfer:{id}:status:{status}`
- **Storage:** In-memory map (development)
- **Behavior:** Duplicate status updates return cached results

### Example

```bash
# First request
curl -X PATCH http://localhost:3400/v1/transfers/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":"completed"}'

# Duplicate request (returns same result, no re-processing)
curl -X PATCH http://localhost:3400/v1/transfers/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":"completed"}'
```

## Testing

The project includes comprehensive unit tests, integration tests, and end-to-end tests following clean architecture principles.

### Test Structure

- **Unit Tests**: Test individual layers in isolation with mocked dependencies
  - Repository tests: `internal/adapters/framework/secondary/repository/*/..._test.go`
  - Core layer tests: `internal/adapters/core/transfer/transfer_test.go`
  - App layer tests: `internal/adapters/app/transfer/transfer_test.go`
  - Handler tests: `internal/adapters/framework/primary/rest_fiber/transfer/handler_test.go`

- **Integration Tests**: Test multiple layers working together
  - `test/integration/transfer_integration_test.go`

- **End-to-End Tests**: Test full HTTP request/response cycle
  - `test/e2e/transfer_e2e_test.go`

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific layer tests
go test ./internal/adapters/core/transfer/... -v
go test ./internal/adapters/app/transfer/... -v
go test ./test/integration/... -v

# Run with coverage
go test ./... -cover

# Use test runner script
./run_tests.sh
```

### Test Coverage

- Repository layer: 100% (5 tests)
- Cache layer: 100% (7 tests)
- Core layer: 100% (3 tests)
- App layer: 100% (5 tests)
- Integration: 100% (10 tests)
- **Total: 30+ tests**

## Development

### Project Structure

- `main.go` - Application entry point and routing
- `handler/handler.go` - HTTP request handlers
- `service/service.go` - Business logic and in-memory repository

### Adding New Features

1. Define interface in `service/service.go`
2. Implement in repository (InMemoryRepository)
3. Add handler in `handler/handler.go`
4. Register route in `main.go`

## Layer Structure

### Entities (`internal/adapters/core/entity/`)
Pure domain models with no business logic.

### Ports (`internal/ports/`)
Interfaces that define contracts between layers:
- **Core Ports**: Business logic interfaces
- **App Ports**: Use case interfaces  
- **Secondary Ports**: External dependency interfaces (DB, cache)

### Core Layer (`internal/adapters/core/`)
Business logic and domain rules. Implements core ports.

### Application Layer (`internal/adapters/app/`)
Use cases and application-specific logic. Orchestrates core layer.

### Framework Adapters (`internal/adapters/framework/`)
- **Primary**: HTTP handlers (incoming requests)
- **Secondary**: Database and cache implementations (outgoing)

## Dependency Flow

```
HTTP Request → Handler → App Layer → Core Layer → Repository → Database
```

**Key Principle**: Dependencies point inward. Core layer has no external dependencies.

## Configuration

Currently uses hardcoded configuration. For production, use environment variables:

```bash
export PORT=3400
export REDIS_URL=localhost:6379
export DATABASE_URL=postgres://user:pass@localhost:5432/transfers
```

## Monitoring

Add metrics for:
- Request latency
- Idempotency hit/miss rate
- Transfer creation rate
- Error rates

See [docs/SCALING_EXAMPLES.md](docs/SCALING_EXAMPLES.md) for Prometheus integration.

## Production Checklist

- [ ] Replace in-memory storage with Redis/PostgreSQL
- [ ] Add environment-based configuration
- [ ] Implement proper error handling
- [ ] Add request validation
- [ ] Set up logging (structured logs)
- [ ] Add metrics and monitoring
- [ ] Implement rate limiting
- [ ] Add authentication/authorization
- [ ] Set up CI/CD pipeline
- [ ] Configure graceful shutdown
- [ ] Add health check endpoints
- [ ] Document API with OpenAPI/Swagger

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

[Add your license here]

## Contact

[Add contact information]
