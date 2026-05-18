#!/bin/bash

set -e

echo "================================"
echo "Transfer Service Test Suite"
echo "================================"
echo ""

echo "1. Running Unit Tests (Repository Layer)..."
go test ./internal/adapters/framework/secondary/repository/db/transfer/... -v
echo "✓ Repository tests passed"
echo ""

echo "2. Running Unit Tests (Cache Layer)..."
go test ./internal/adapters/framework/secondary/repository/cache/idempotency/... -v
echo "✓ Cache tests passed"
echo ""

echo "3. Running Unit Tests (Core Layer)..."
go test ./internal/adapters/core/transfer/... -v
echo "✓ Core layer tests passed"
echo ""

echo "4. Running Unit Tests (App Layer)..."
go test ./internal/adapters/app/transfer/... -v
echo "✓ App layer tests passed"
echo ""

echo "5. Running Integration Tests..."
go test ./test/integration/... -v
echo "✓ Integration tests passed"
echo ""

echo "6. Running All Tests with Coverage..."
go test ./... -cover -v 2>&1 | grep -E "(coverage|PASS|FAIL)" | tail -20
echo ""

echo "================================"
echo "All Tests Passed! ✓"
echo "================================"
