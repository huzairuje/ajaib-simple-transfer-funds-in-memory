package transfer

import (
	"context"
	"fmt"
	"testing"
	"ajaib-testing-code/internal/adapters/core/entity"
)

type MockDB struct {
	transfers map[int64]entity.Transfer
	nextID    int64
}

func NewMockDB() *MockDB {
	return &MockDB{
		transfers: make(map[int64]entity.Transfer),
		nextID:    1,
	}
}

func (m *MockDB) CreateTransfer(ctx context.Context, t entity.Transfer) (int64, error) {
	t.ID = m.nextID
	m.transfers[m.nextID] = t
	m.nextID++
	return t.ID, nil
}

func (m *MockDB) GetTransferByID(ctx context.Context, id int64) (*entity.Transfer, error) {
	if t, exists := m.transfers[id]; exists {
		return &t, nil
	}
	return nil, nil
}

func (m *MockDB) GetListTransfer(ctx context.Context) ([]entity.Transfer, error) {
	var result []entity.Transfer
	for _, t := range m.transfers {
		result = append(result, t)
	}
	return result, nil
}

func (m *MockDB) UpdateTransferStatus(ctx context.Context, id int64, status string) error {
	if t, exists := m.transfers[id]; exists {
		t.Status = status
		m.transfers[id] = t
		return nil
	}
	return nil
}

type MockCache struct {
	records map[string]entity.IdempotencyRecord
}

func NewMockCache() *MockCache {
	return &MockCache{
		records: make(map[string]entity.IdempotencyRecord),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) (*entity.IdempotencyRecord, error) {
	if record, exists := m.records[key]; exists {
		return &record, nil
	}
	return nil, fmt.Errorf("key not found")
}

func (m *MockCache) Set(ctx context.Context, key string, record entity.IdempotencyRecord) error {
	m.records[key] = record
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) bool {
	_, exists := m.records[key]
	return exists
}

func TestTransferCore_CreateTransfer(t *testing.T) {
	mockDB := NewMockDB()
	mockCache := NewMockCache()
	core := New(Config{DB: mockDB, Cache: mockCache})
	ctx := context.Background()

	transfer := entity.Transfer{
		FromAccountID: 1001,
		ToAccountID:   1002,
		Amount:        50000,
		Currency:      "IDR",
		Status:        "success",
	}

	id, err := core.CreateTransfer(ctx, transfer)
	if err != nil {
		t.Fatalf("CreateTransfer failed: %v", err)
	}

	if id != 1 {
		t.Errorf("Expected ID 1, got %d", id)
	}
}

func TestTransferCore_UpdateTransferStatus_Idempotent(t *testing.T) {
	mockDB := NewMockDB()
	mockCache := NewMockCache()
	core := New(Config{DB: mockDB, Cache: mockCache})
	ctx := context.Background()

	transfer := entity.Transfer{
		FromAccountID: 1001,
		ToAccountID:   1002,
		Amount:        50000,
		Currency:      "IDR",
		Status:        "pending",
	}

	id, _ := core.CreateTransfer(ctx, transfer)
	key := "transfer:1:status:completed"

	result1, err := core.UpdateTransferStatus(ctx, id, "completed", key)
	if err != nil {
		t.Fatalf("First UpdateTransferStatus failed: %v", err)
	}

	if result1.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result1.Status)
	}

	result2, err := core.UpdateTransferStatus(ctx, id, "completed", key)
	if err != nil {
		t.Fatalf("Second UpdateTransferStatus failed: %v", err)
	}

	if result2.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result2.Status)
	}
}

func TestTransferCore_UpdateTransferStatus_Conflict(t *testing.T) {
	mockDB := NewMockDB()
	mockCache := NewMockCache()
	core := New(Config{DB: mockDB, Cache: mockCache})
	ctx := context.Background()

	transfer := entity.Transfer{
		FromAccountID: 1001,
		ToAccountID:   1002,
		Amount:        50000,
		Currency:      "IDR",
		Status:        "pending",
	}

	id, _ := core.CreateTransfer(ctx, transfer)
	key := "transfer:1:status:completed"

	core.UpdateTransferStatus(ctx, id, "completed", key)

	_, err := core.UpdateTransferStatus(ctx, id, "failed", key)
	if err == nil {
		t.Error("Expected conflict error, got nil")
	}
}
