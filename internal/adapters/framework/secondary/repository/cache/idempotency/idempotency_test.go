package idempotency

import (
	"context"
	"fmt"
	"testing"
	"ajaib-testing-code/internal/adapters/core/entity"
)

func TestInMemoryCache_Set_Get(t *testing.T) {
	cache := New(Config{})
	ctx := context.Background()

	record := entity.IdempotencyRecord{
		Key:        "transfer:1:status:completed",
		TransferID: 1,
		Status:     "completed",
	}

	err := cache.Set(ctx, record.Key, record)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := cache.Get(ctx, record.Key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.TransferID != record.TransferID {
		t.Errorf("Expected TransferID %d, got %d", record.TransferID, retrieved.TransferID)
	}
	if retrieved.Status != record.Status {
		t.Errorf("Expected Status '%s', got '%s'", record.Status, retrieved.Status)
	}
}

func TestInMemoryCache_Get_NotFound(t *testing.T) {
	cache := New(Config{})
	ctx := context.Background()

	_, err := cache.Get(ctx, "non-existing-key")
	if err == nil {
		t.Error("Expected error for non-existing key, got nil")
	}
}

func TestInMemoryCache_Exists(t *testing.T) {
	cache := New(Config{})
	ctx := context.Background()

	key := "transfer:1:status:completed"
	record := entity.IdempotencyRecord{
		Key:        key,
		TransferID: 1,
		Status:     "completed",
	}

	t.Run("key does not exist", func(t *testing.T) {
		if cache.Exists(ctx, key) {
			t.Error("Expected key to not exist")
		}
	})

	cache.Set(ctx, key, record)

	t.Run("key exists", func(t *testing.T) {
		if !cache.Exists(ctx, key) {
			t.Error("Expected key to exist")
		}
	})
}

func TestInMemoryCache_MultipleRecords(t *testing.T) {
	cache := New(Config{})
	ctx := context.Background()

	records := []entity.IdempotencyRecord{
		{Key: "transfer:1:status:completed", TransferID: 1, Status: "completed"},
		{Key: "transfer:2:status:pending", TransferID: 2, Status: "pending"},
		{Key: "transfer:3:status:failed", TransferID: 3, Status: "failed"},
	}

	for _, record := range records {
		cache.Set(ctx, record.Key, record)
	}

	for _, record := range records {
		retrieved, err := cache.Get(ctx, record.Key)
		if err != nil {
			t.Fatalf("Get failed for key %s: %v", record.Key, err)
		}
		if retrieved.TransferID != record.TransferID {
			t.Errorf("Expected TransferID %d, got %d", record.TransferID, retrieved.TransferID)
		}
	}
}

func TestInMemoryCache_Concurrency(t *testing.T) {
	cache := New(Config{})
	ctx := context.Background()

	done := make(chan bool)
	for i := 1; i <= 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("transfer:%d:status:completed", index)
			record := entity.IdempotencyRecord{
				Key:        key,
				TransferID: int64(index),
				Status:     "completed",
			}
			cache.Set(ctx, key, record)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if !cache.Exists(ctx, "transfer:1:status:completed") {
		t.Error("Expected key to exist after concurrent writes")
	}
}
