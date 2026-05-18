package transfer

import (
	"context"
	"testing"
	"ajaib-testing-code/internal/adapters/core/entity"
)

func TestInMemoryRepository_CreateTransfer(t *testing.T) {
	repo := New(Config{})
	ctx := context.Background()

	transfer := entity.Transfer{
		FromAccountID: 1001,
		ToAccountID:   1002,
		Amount:        50000,
		Currency:      "IDR",
		Status:        "success",
		FromBalance:   100000,
		ToBalance:     50000,
	}

	id, err := repo.CreateTransfer(ctx, transfer)
	if err != nil {
		t.Fatalf("CreateTransfer failed: %v", err)
	}

	if id != 1 {
		t.Errorf("Expected ID 1, got %d", id)
	}

	created, err := repo.GetTransferByID(ctx, id)
	if err != nil {
		t.Fatalf("GetTransferByID failed: %v", err)
	}

	if created.FromAccountID != transfer.FromAccountID {
		t.Errorf("Expected FromAccountID %d, got %d", transfer.FromAccountID, created.FromAccountID)
	}
	if created.Amount != transfer.Amount {
		t.Errorf("Expected Amount %d, got %d", transfer.Amount, created.Amount)
	}
}

func TestInMemoryRepository_GetTransferByID(t *testing.T) {
	repo := New(Config{})
	ctx := context.Background()

	transfer := entity.Transfer{
		FromAccountID: 1001,
		ToAccountID:   1002,
		Amount:        50000,
		Currency:      "IDR",
		Status:        "success",
	}

	id, _ := repo.CreateTransfer(ctx, transfer)

	t.Run("existing transfer", func(t *testing.T) {
		result, err := repo.GetTransferByID(ctx, id)
		if err != nil {
			t.Fatalf("GetTransferByID failed: %v", err)
		}
		if result.ID != id {
			t.Errorf("Expected ID %d, got %d", id, result.ID)
		}
	})

	t.Run("non-existing transfer", func(t *testing.T) {
		_, err := repo.GetTransferByID(ctx, 999)
		if err == nil {
			t.Error("Expected error for non-existing transfer, got nil")
		}
	})
}

func TestInMemoryRepository_GetListTransfer(t *testing.T) {
	repo := New(Config{})
	ctx := context.Background()

	transfers := []entity.Transfer{
		{FromAccountID: 1001, ToAccountID: 1002, Amount: 50000, Currency: "IDR", Status: "success"},
		{FromAccountID: 1003, ToAccountID: 1004, Amount: 75000, Currency: "IDR", Status: "success"},
		{FromAccountID: 1005, ToAccountID: 1006, Amount: 100000, Currency: "IDR", Status: "pending"},
	}

	for _, transfer := range transfers {
		_, err := repo.CreateTransfer(ctx, transfer)
		if err != nil {
			t.Fatalf("CreateTransfer failed: %v", err)
		}
	}

	list, err := repo.GetListTransfer(ctx)
	if err != nil {
		t.Fatalf("GetListTransfer failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 transfers, got %d", len(list))
	}
}

func TestInMemoryRepository_UpdateTransferStatus(t *testing.T) {
	repo := New(Config{})
	ctx := context.Background()

	transfer := entity.Transfer{
		FromAccountID: 1001,
		ToAccountID:   1002,
		Amount:        50000,
		Currency:      "IDR",
		Status:        "pending",
	}

	id, _ := repo.CreateTransfer(ctx, transfer)

	t.Run("update existing transfer", func(t *testing.T) {
		err := repo.UpdateTransferStatus(ctx, id, "completed")
		if err != nil {
			t.Fatalf("UpdateTransferStatus failed: %v", err)
		}

		updated, _ := repo.GetTransferByID(ctx, id)
		if updated.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", updated.Status)
		}
	})

	t.Run("update non-existing transfer", func(t *testing.T) {
		err := repo.UpdateTransferStatus(ctx, 999, "completed")
		if err == nil {
			t.Error("Expected error for non-existing transfer, got nil")
		}
	})
}

func TestInMemoryRepository_Concurrency(t *testing.T) {
	repo := New(Config{})
	ctx := context.Background()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			transfer := entity.Transfer{
				FromAccountID: int64(1000 + index),
				ToAccountID:   int64(2000 + index),
				Amount:        int64(10000 * index),
				Currency:      "IDR",
				Status:        "success",
			}
			_, err := repo.CreateTransfer(ctx, transfer)
			if err != nil {
				t.Errorf("Concurrent CreateTransfer failed: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	list, _ := repo.GetListTransfer(ctx)
	if len(list) != 10 {
		t.Errorf("Expected 10 transfers after concurrent writes, got %d", len(list))
	}
}
