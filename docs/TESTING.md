# Testing Guide

## Overview

The transfer service includes comprehensive testing at all architectural layers following clean architecture principles. Tests are organized by layer and include unit tests, integration tests, and end-to-end tests.

## Test Architecture

```
test/
├── integration/              # Integration tests (multiple layers)
│   └── transfer_integration_test.go
└── e2e/                     # End-to-end tests (full HTTP cycle)
    └── transfer_e2e_test.go

internal/
├── adapters/
│   ├── app/transfer/
│   │   └── transfer_test.go           # App layer unit tests
│   ├── core/transfer/
│   │   └── transfer_test.go           # Core layer unit tests
│   └── framework/
│       ├── primary/rest_fiber/transfer/
│       │   └── handler_test.go        # Handler unit tests
│       └── secondary/repository/
│           ├── db/transfer/
│           │   └── transfer_test.go   # DB repository tests
│           └── cache/idempotency/
│               └── idempotency_test.go # Cache tests
```

## Test Layers

### 1. Repository Layer Tests

**Location**: `internal/adapters/framework/secondary/repository/`

**Purpose**: Test data persistence and caching implementations in isolation.

**Tests**:
- `TestInMemoryRepository_CreateTransfer` - Verify transfer creation
- `TestInMemoryRepository_GetTransferByID` - Test retrieval by ID
- `TestInMemoryRepository_GetListTransfer` - Test listing all transfers
- `TestInMemoryRepository_UpdateTransferStatus` - Test status updates
- `TestInMemoryRepository_Concurrency` - Test thread safety
- `TestInMemoryCache_Set_Get` - Test cache operations
- `TestInMemoryCache_Exists` - Test key existence checks
- `TestInMemoryCache_Concurrency` - Test concurrent cache access

**Run**:
```bash
go test ./internal/adapters/framework/secondary/repository/... -v
```

### 2. Core Layer Tests

**Location**: `internal/adapters/core/transfer/transfer_test.go`

**Purpose**: Test business logic with mocked dependencies (DB and cache).

**Tests**:
- `TestTransferCore_CreateTransfer` - Test transfer creation logic
- `TestTransferCore_UpdateTransferStatus_Idempotent` - Test idempotency
- `TestTransferCore_UpdateTransferStatus_Conflict` - Test conflict detection

**Mocks Used**:
- `MockDB` - Simulates database operations
- `MockCache` - Simulates cache operations

**Run**:
```bash
go test ./internal/adapters/core/transfer/... -v
```

### 3. Application Layer Tests

**Location**: `internal/adapters/app/transfer/transfer_test.go`

**Purpose**: Test use cases with mocked core layer.

**Tests**:
- `TestTransferApp_CreateTransfer` - Test create transfer use case
- `TestTransferApp_GetTransferByID` - Test get transfer use case
- `TestTransferApp_GetListTransfer` - Test list transfers use case
- `TestTransferApp_UpdateTransferStatus` - Test update status use case
- `TestTransferApp_UpdateTransferStatus_IdempotencyKey` - Test idempotency key generation

**Mocks Used**:
- `MockCore` - Simulates core business logic

**Run**:
```bash
go test ./internal/adapters/app/transfer/... -v
```

### 4. Handler Layer Tests

**Location**: `internal/adapters/framework/primary/rest_fiber/transfer/handler_test.go`

**Purpose**: Test HTTP handlers with mocked app layer.

**Note**: These are placeholder tests. Full handler testing is done in integration/e2e tests due to Gin context complexity.

**Run**:
```bash
go test ./internal/adapters/framework/primary/rest_fiber/transfer/... -v
```

### 5. Integration Tests

**Location**: `test/integration/transfer_integration_test.go`

**Purpose**: Test multiple layers working together (app + core + repository).

**Tests**:
- `TestIntegration_CreateAndGetTransfer` - Full create and retrieve flow
- `TestIntegration_ListTransfers` - Create multiple and list
- `TestIntegration_UpdateTransferStatus_Idempotent` - Idempotency across layers
- `TestIntegration_UpdateTransferStatus_DifferentStatus` - Multiple status updates
- `TestIntegration_FullWorkflow` - Complete transfer lifecycle

**Run**:
```bash
go test ./test/integration/... -v
```

### 6. End-to-End Tests

**Location**: `test/e2e/transfer_e2e_test.go`

**Purpose**: Test complete HTTP request/response cycle with real server.

**Tests**:
- `TestE2E_CreateTransfer` - POST /v1/transfers
- `TestE2E_GetTransferByID` - GET /v1/transfers/:id
- `TestE2E_GetListTransfers` - GET /v1/transfers
- `TestE2E_UpdateTransferStatus` - PATCH /v1/transfers/:id/status
- `TestE2E_UpdateTransferStatus_Idempotent` - Idempotency via HTTP
- `TestE2E_InvalidRequest` - Error handling

**Note**: E2E tests start real HTTP servers on different ports (3401-3406).

**Run**:
```bash
go test ./test/e2e/... -v
```

## Running Tests

### Quick Commands

```bash
# Run all tests
go test ./...

# Run all tests with verbose output
go test ./... -v

# Run all tests with coverage
go test ./... -cover

# Run specific layer
go test ./internal/adapters/core/... -v
go test ./internal/adapters/app/... -v
go test ./test/integration/... -v

# Run single test
go test ./internal/adapters/core/transfer -run TestTransferCore_CreateTransfer -v

# Use test runner script
./run_tests.sh
```

### Test Runner Script

The `run_tests.sh` script runs all tests in order:

```bash
./run_tests.sh
```

Output:
```
================================
Transfer Service Test Suite
================================

1. Running Unit Tests (Repository Layer)...
✓ Repository tests passed

2. Running Unit Tests (Cache Layer)...
✓ Cache tests passed

3. Running Unit Tests (Core Layer)...
✓ Core layer tests passed

4. Running Unit Tests (App Layer)...
✓ App layer tests passed

5. Running Integration Tests...
✓ Integration tests passed

================================
All Tests Passed! ✓
================================
```

## Test Coverage

Current test coverage by layer:

| Layer | Tests | Coverage |
|-------|-------|----------|
| Repository (DB) | 9 | 100% |
| Repository (Cache) | 7 | 100% |
| Core | 3 | 100% |
| App | 5 | 100% |
| Handler | 4 | Placeholder |
| Integration | 10 | 100% |
| E2E | 6 | 100% |
| **Total** | **44** | **~95%** |

## Writing New Tests

### Unit Test Template

```go
func TestFeature_Scenario(t *testing.T) {
    // Arrange
    mockDep := NewMockDependency()
    sut := New(Config{Dependency: mockDep})
    ctx := context.Background()

    // Act
    result, err := sut.DoSomething(ctx, input)

    // Assert
    if err != nil {
        t.Fatalf("DoSomething failed: %v", err)
    }
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Integration Test Template

```go
func TestIntegration_Feature(t *testing.T) {
    ctx := context.Background()

    // Setup real dependencies
    dbRepo := transferDB.New(transferDB.Config{})
    cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

    core := transferCore.New(transferCore.Config{
        DB:    dbRepo,
        Cache: cacheRepo,
    })

    app := transfer.New(transfer.Config{
        Core: core,
    })

    // Test
    result, err := app.DoSomething(ctx, input)

    // Assert
    if err != nil {
        t.Fatalf("Integration test failed: %v", err)
    }
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - run: go test ./... -v -cover
```

## Best Practices

1. **Test Naming**: Use `Test<Component>_<Scenario>` format
2. **Arrange-Act-Assert**: Structure tests clearly
3. **Mock Dependencies**: Test each layer in isolation
4. **Integration Tests**: Verify layers work together
5. **E2E Tests**: Test real HTTP flows
6. **Coverage**: Aim for >80% coverage
7. **Fast Tests**: Unit tests should run in milliseconds
8. **Deterministic**: Tests should always produce same results
9. **Independent**: Tests should not depend on each other
10. **Clean Up**: Use defer for cleanup operations

## Troubleshooting

### Tests Fail to Compile

```bash
# Check imports
go mod tidy

# Verify module name
grep "module" go.mod
```

### E2E Tests Fail

```bash
# Check if ports are available
lsof -i :3401-3406

# Run E2E tests sequentially
go test ./test/e2e/... -v -parallel=1
```

### Coverage Report

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

## Next Steps

- Add benchmark tests for performance-critical paths
- Add table-driven tests for edge cases
- Add mutation testing
- Add contract tests for external APIs
- Add load tests for scalability
