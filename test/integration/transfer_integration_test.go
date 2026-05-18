package integration

import (
	"context"
	"testing"
	"ajaib-testing-code/internal/adapters/app/transfer"
	"ajaib-testing-code/internal/adapters/core/entity"
	transferCore "ajaib-testing-code/internal/adapters/core/transfer"
	idempotencyCache "ajaib-testing-code/internal/adapters/framework/secondary/repository/cache/idempotency"
	transferDB "ajaib-testing-code/internal/adapters/framework/secondary/repository/db/transfer"
)

func TestIntegration_CreateAndGetTransfer(t *testing.T) {
	ctx := context.Background()

	dbRepo := transferDB.New(transferDB.Config{})
	cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

	core := transferCore.New(transferCore.Config{
		DB:    dbRepo,
		Cache: cacheRepo,
	})

	app := transfer.New(transfer.Config{
		Core: core,
	})

	request := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	id, err := app.CreateTransfer(ctx, request)
	if err != nil {
		t.Fatalf("CreateTransfer failed: %v", err)
	}

	result, err := app.GetTransferByID(ctx, id)
	if err != nil {
		t.Fatalf("GetTransferByID failed: %v", err)
	}

	if result.FromAccountID != request.From {
		t.Errorf("Expected FromAccountID %d, got %d", request.From, result.FromAccountID)
	}
	if result.Amount != request.Amount {
		t.Errorf("Expected Amount %d, got %d", request.Amount, result.Amount)
	}
	if result.Status != "success" {
		t.Errorf("Expected Status 'success', got '%s'", result.Status)
	}
}

func TestIntegration_ListTransfers(t *testing.T) {
	ctx := context.Background()

	dbRepo := transferDB.New(transferDB.Config{})
	cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

	core := transferCore.New(transferCore.Config{
		DB:    dbRepo,
		Cache: cacheRepo,
	})

	app := transfer.New(transfer.Config{
		Core: core,
	})

	requests := []entity.CreateTransferRequest{
		{From: 1001, To: 1002, Amount: 50000, Currency: "IDR", FromBalance: 100000, ToBalance: 50000},
		{From: 1003, To: 1004, Amount: 75000, Currency: "IDR", FromBalance: 150000, ToBalance: 75000},
		{From: 1005, To: 1006, Amount: 100000, Currency: "IDR", FromBalance: 200000, ToBalance: 100000},
	}

	for _, req := range requests {
		_, err := app.CreateTransfer(ctx, req)
		if err != nil {
			t.Fatalf("CreateTransfer failed: %v", err)
		}
	}

	list, err := app.GetListTransfer(ctx)
	if err != nil {
		t.Fatalf("GetListTransfer failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 transfers, got %d", len(list))
	}
}

func TestIntegration_UpdateTransferStatus_Idempotent(t *testing.T) {
	ctx := context.Background()

	dbRepo := transferDB.New(transferDB.Config{})
	cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

	core := transferCore.New(transferCore.Config{
		DB:    dbRepo,
		Cache: cacheRepo,
	})

	app := transfer.New(transfer.Config{
		Core: core,
	})

	request := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	id, _ := app.CreateTransfer(ctx, request)

	result1, err := app.UpdateTransferStatus(ctx, id, "completed")
	if err != nil {
		t.Fatalf("First UpdateTransferStatus failed: %v", err)
	}

	result2, err := app.UpdateTransferStatus(ctx, id, "completed")
	if err != nil {
		t.Fatalf("Second UpdateTransferStatus failed: %v", err)
	}

	if result1.Status != result2.Status {
		t.Error("Idempotent calls should return same status")
	}

	if result1.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result1.Status)
	}
}

func TestIntegration_UpdateTransferStatus_DifferentStatus(t *testing.T) {
	ctx := context.Background()

	dbRepo := transferDB.New(transferDB.Config{})
	cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

	core := transferCore.New(transferCore.Config{
		DB:    dbRepo,
		Cache: cacheRepo,
	})

	app := transfer.New(transfer.Config{
		Core: core,
	})

	request := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	id, _ := app.CreateTransfer(ctx, request)

	result1, _ := app.UpdateTransferStatus(ctx, id, "processing")
	if result1.Status != "processing" {
		t.Errorf("Expected status 'processing', got '%s'", result1.Status)
	}

	result2, _ := app.UpdateTransferStatus(ctx, id, "completed")
	if result2.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result2.Status)
	}
}

func TestIntegration_FullWorkflow(t *testing.T) {
	ctx := context.Background()

	dbRepo := transferDB.New(transferDB.Config{})
	cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

	core := transferCore.New(transferCore.Config{
		DB:    dbRepo,
		Cache: cacheRepo,
	})

	app := transfer.New(transfer.Config{
		Core: core,
	})

	t.Run("create transfer", func(t *testing.T) {
		request := entity.CreateTransferRequest{
			From:        1001,
			To:          1002,
			Amount:      50000,
			Currency:    "IDR",
			FromBalance: 100000,
			ToBalance:   50000,
		}

		id, err := app.CreateTransfer(ctx, request)
		if err != nil {
			t.Fatalf("CreateTransfer failed: %v", err)
		}
		if id != 1 {
			t.Errorf("Expected ID 1, got %d", id)
		}
	})

	t.Run("get transfer details", func(t *testing.T) {
		result, err := app.GetTransferByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetTransferByID failed: %v", err)
		}
		if result.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", result.Status)
		}
	})

	t.Run("update status to processing", func(t *testing.T) {
		result, err := app.UpdateTransferStatus(ctx, 1, "processing")
		if err != nil {
			t.Fatalf("UpdateTransferStatus failed: %v", err)
		}
		if result.Status != "processing" {
			t.Errorf("Expected status 'processing', got '%s'", result.Status)
		}
	})

	t.Run("update status to completed", func(t *testing.T) {
		result, err := app.UpdateTransferStatus(ctx, 1, "completed")
		if err != nil {
			t.Fatalf("UpdateTransferStatus failed: %v", err)
		}
		if result.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", result.Status)
		}
	})

	t.Run("verify final state", func(t *testing.T) {
		result, err := app.GetTransferByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetTransferByID failed: %v", err)
		}
		if result.Status != "completed" {
			t.Errorf("Expected final status 'completed', got '%s'", result.Status)
		}
	})
}
