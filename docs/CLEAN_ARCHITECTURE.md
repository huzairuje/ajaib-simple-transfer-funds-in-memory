# Clean Architecture Implementation

## Overview

The transfer service has been refactored to follow Clean Architecture principles based on the knowledge-service reference project. This architecture separates concerns into distinct layers, making the codebase more maintainable, testable, and scalable.

## Architecture Layers

```
cmd/gateway/                          # Application entry point
├── main.go                          # Dependency injection & initialization

internal/
├── adapters/
│   ├── app/                         # Application Layer (Use Cases)
│   │   └── transfer/
│   │       └── transfer.go          # Use case implementations
│   │
│   ├── core/                        # Core/Domain Layer (Business Logic)
│   │   ├── entity/
│   │   │   └── transfer.go          # Domain entities
│   │   └── transfer/
│   │       └── transfer.go          # Business logic implementations
│   │
│   └── framework/                   # Framework Adapters
│       ├── primary/                 # Incoming Adapters
│       │   └── rest_fiber/
│       │       └── transfer/
│       │           └── handler.go   # HTTP handlers
│       │
│       └── secondary/               # Outgoing Adapters
│           └── repository/
│               ├── db/
│               │   └── transfer/
│               │       └── transfer.go    # Database implementation
│               └── cache/
│                   └── idempotency/
│                       └── idempotency.go # Cache implementation
│
└── ports/                           # Interfaces/Contracts
    ├── app/
    │   └── transfer.go              # Application port interfaces
    ├── core/
    │   └── transfer.go              # Core port interfaces
    └── secondary/
        ├── db/
        │   └── transfer.go          # Database port interfaces
        └── cache/
            └── idempotency.go       # Cache port interfaces

config/
└── config.go                        # Configuration management

router/
└── router.go                        # HTTP routing setup
```

## Layer Responsibilities

### 1. Entities (`internal/adapters/core/entity/`)

Domain models and data structures. These are pure data structures with no business logic.

**Example:**
```go
type Transfer struct {
    ID            int64
    FromAccountID int64
    ToAccountID   int64
    Amount        int64
    Currency      string
    Status        string
}
```

### 2. Ports (`internal/ports/`)

Interfaces that define contracts between layers. This enables dependency inversion.

**Types:**
- **Core Ports**: Interfaces for business logic
- **App Ports**: Interfaces for use cases
- **Secondary Ports**: Interfaces for external dependencies (DB, cache, APIs)

### 3. Core Layer (`internal/adapters/core/`)

Contains business logic and domain rules. Depends only on entities and ports.

**Responsibilities:**
- Implement business rules
- Validate domain constraints
- Coordinate between repositories
- Handle idempotency logic

### 4. Application Layer (`internal/adapters/app/`)

Contains use cases and application-specific logic. Orchestrates the core layer.

**Responsibilities:**
- Transform requests to domain models
- Call core layer methods
- Handle application-level concerns
- Generate idempotency keys

### 5. Framework Adapters (`internal/adapters/framework/`)

#### Primary Adapters (Incoming)
Handle external requests (HTTP, gRPC, MQ).

**Responsibilities:**
- Parse HTTP requests
- Validate input
- Call application layer
- Format responses
- Handle HTTP status codes

#### Secondary Adapters (Outgoing)
Implement external dependencies (DB, cache, APIs).

**Responsibilities:**
- Implement repository interfaces
- Handle data persistence
- Manage connections
- Handle errors

## Dependency Flow

```
HTTP Request
    ↓
Handler (Primary Adapter)
    ↓
Application Layer (Use Case)
    ↓
Core Layer (Business Logic)
    ↓
Repository (Secondary Adapter)
    ↓
Database/Cache
```

**Key Principle:** Dependencies point inward. Core layer has no dependencies on outer layers.

## Running the Application

### Development

```bash
# Run from project root
go run cmd/gateway/main.go

# Or build and run
go build -o bin/transfer-service cmd/gateway/main.go
./bin/transfer-service
```

### Environment Variables

```bash
export PORT=3400
export APP_NAME=transfer-service
```

### API Endpoints

Same as before:
- `POST /v1/transfers` - Create transfer
- `GET /v1/transfers` - List transfers
- `GET /v1/transfers/:id` - Get transfer details
- `PATCH /v1/transfers/:id/status` - Update transfer status (idempotent)

## Adding New Features

### Example: Add Transfer Cancellation

#### 1. Update Entity
```go
// internal/adapters/core/entity/transfer.go
type CancelTransferRequest struct {
    Reason string `json:"reason" binding:"required"`
}
```

#### 2. Update Ports
```go
// internal/ports/core/transfer.go
type TransferInterface interface {
    // ... existing methods
    CancelTransfer(ctx context.Context, id int64, reason string) error
}

// internal/ports/app/transfer.go
type TransferInterface interface {
    // ... existing methods
    CancelTransfer(ctx context.Context, id int64, request entity.CancelTransferRequest) error
}
```

#### 3. Implement Core Logic
```go
// internal/adapters/core/transfer/cancel_transfer.go
func (t *transfer) CancelTransfer(ctx context.Context, id int64, reason string) error {
    // Business logic here
    return t.db.CancelTransfer(ctx, id, reason)
}
```

#### 4. Implement Use Case
```go
// internal/adapters/app/transfer/cancel_transfer.go
func (t *transfer) CancelTransfer(ctx context.Context, id int64, request entity.CancelTransferRequest) error {
    return t.core.CancelTransfer(ctx, id, request.Reason)
}
```

#### 5. Add Handler
```go
// internal/adapters/framework/primary/rest_fiber/transfer/handler.go
func (h *Handler) CancelTransferHandler(c *gin.Context) {
    // Parse request, call app layer, return response
}
```

#### 6. Register Route
```go
// router/router.go
transfers.DELETE("/:id", config.TransferHandler.CancelTransferHandler)
```

## Benefits of Clean Architecture

### 1. Testability
Each layer can be tested independently with mocks.

```go
// Test core layer with mock repository
mockDB := &MockTransferDB{}
core := transferCore.New(transferCore.Config{DB: mockDB})
```

### 2. Maintainability
Clear separation of concerns makes code easier to understand and modify.

### 3. Flexibility
Easy to swap implementations (e.g., in-memory → PostgreSQL).

```go
// Switch from in-memory to PostgreSQL
dbRepo := postgresDB.New(postgresDB.Config{...})
```

### 4. Scalability
Add new features without affecting existing code.

### 5. Independence
Business logic is independent of frameworks, databases, and external services.

## Migration from Old Structure

### Old Structure
```
handler/
  handler.go          # HTTP handlers + business logic
service/
  service.go          # Repository + business logic mixed
main.go               # Entry point
```

### New Structure
- **handler.go** → Split into:
  - `internal/adapters/framework/primary/rest_fiber/transfer/handler.go` (HTTP handling)
  - `internal/adapters/app/transfer/transfer.go` (Use cases)
  - `internal/adapters/core/transfer/transfer.go` (Business logic)

- **service.go** → Split into:
  - `internal/adapters/core/entity/transfer.go` (Entities)
  - `internal/adapters/framework/secondary/repository/db/transfer/transfer.go` (Repository)
  - `internal/ports/secondary/db/transfer.go` (Repository interface)

## Testing Strategy

### Unit Tests
Test each layer independently:

```go
// Test core layer
func TestCreateTransfer(t *testing.T) {
    mockDB := &MockTransferDB{}
    mockCache := &MockIdempotencyCache{}
    core := transferCore.New(transferCore.Config{
        DB: mockDB,
        Cache: mockCache,
    })
    // Test business logic
}
```

### Integration Tests
Test multiple layers together:

```go
// Test app + core layers
func TestTransferUseCase(t *testing.T) {
    // Setup real or mock dependencies
    // Test complete use case flow
}
```

## Next Steps

1. **Add Redis Cache**: Replace in-memory cache with Redis
2. **Add PostgreSQL**: Replace in-memory DB with PostgreSQL
3. **Add Validation**: Add input validation middleware
4. **Add Metrics**: Add Prometheus metrics
5. **Add Tracing**: Add OpenTelemetry tracing
6. **Add Tests**: Add comprehensive unit and integration tests

## References

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
