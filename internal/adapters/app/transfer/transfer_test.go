package transfer

import (
	"context"
	"testing"
	"ajaib-testing-code/internal/adapters/core/entity"
)

type MockCore struct {
	transfers map[int64]*entity.Transfer
	nextID    int64
}

func NewMockCore() *MockCore {
	return &MockCore{
		transfers: make(map[int64]*entity.Transfer),
		nextID:    1,
	}
}

func (m *MockCore) CreateTransfer(ctx context.Context, t entity.Transfer) (int64, error) {
	t.ID = m.nextID
	m.transfers[m.nextID] = &t
	m.nextID++
	return t.ID, nil
}

func (m *MockCore) GetTransferByID(ctx context.Context, id int64) (*entity.Transfer, error) {
	if t, exists := m.transfers[id]; exists {
		return t, nil
	}
	return nil, nil
}

func (m *MockCore) GetListTransfer(ctx context.Context) ([]entity.Transfer, error) {
	var result []entity.Transfer
	for _, t := range m.transfers {
		result = append(result, *t)
	}
	return result, nil
}

func (m *MockCore) UpdateTransferStatus(ctx context.Context, id int64, status string, key string) (*entity.Transfer, error) {
	if t, exists := m.transfers[id]; exists {
		t.Status = status
		return t, nil
	}
	return nil, nil
}

func TestTransferApp_CreateTransfer(t *testing.T) {
	mockCore := NewMockCore()
	app := New(Config{Core: mockCore})
	ctx := context.Background()

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

	created, _ := mockCore.GetTransferByID(ctx, id)
	if created.FromAccountID != request.From {
		t.Errorf("Expected FromAccountID %d, got %d", request.From, created.FromAccountID)
	}
	if created.Amount != request.Amount {
		t.Errorf("Expected Amount %d, got %d", request.Amount, created.Amount)
	}
	if created.Status != "success" {
		t.Errorf("Expected Status 'success', got '%s'", created.Status)
	}
}

func TestTransferApp_GetTransferByID(t *testing.T) {
	mockCore := NewMockCore()
	app := New(Config{Core: mockCore})
	ctx := context.Background()

	request := entity.CreateTransferRequest{
		From:     1001,
		To:       1002,
		Amount:   50000,
		Currency: "IDR",
	}

	id, _ := app.CreateTransfer(ctx, request)

	result, err := app.GetTransferByID(ctx, id)
	if err != nil {
		t.Fatalf("GetTransferByID failed: %v", err)
	}

	if result.ID != id {
		t.Errorf("Expected ID %d, got %d", id, result.ID)
	}
}

func TestTransferApp_GetListTransfer(t *testing.T) {
	mockCore := NewMockCore()
	app := New(Config{Core: mockCore})
	ctx := context.Background()

	requests := []entity.CreateTransferRequest{
		{From: 1001, To: 1002, Amount: 50000, Currency: "IDR"},
		{From: 1003, To: 1004, Amount: 75000, Currency: "IDR"},
		{From: 1005, To: 1006, Amount: 100000, Currency: "IDR"},
	}

	for _, req := range requests {
		app.CreateTransfer(ctx, req)
	}

	list, err := app.GetListTransfer(ctx)
	if err != nil {
		t.Fatalf("GetListTransfer failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 transfers, got %d", len(list))
	}
}

func TestTransferApp_UpdateTransferStatus(t *testing.T) {
	mockCore := NewMockCore()
	app := New(Config{Core: mockCore})
	ctx := context.Background()

	request := entity.CreateTransferRequest{
		From:     1001,
		To:       1002,
		Amount:   50000,
		Currency: "IDR",
	}

	id, _ := app.CreateTransfer(ctx, request)

	result, err := app.UpdateTransferStatus(ctx, id, "completed")
	if err != nil {
		t.Fatalf("UpdateTransferStatus failed: %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result.Status)
	}
}

func TestTransferApp_UpdateTransferStatus_IdempotencyKey(t *testing.T) {
	mockCore := NewMockCore()
	app := New(Config{Core: mockCore})
	ctx := context.Background()

	request := entity.CreateTransferRequest{
		From:     1001,
		To:       1002,
		Amount:   50000,
		Currency: "IDR",
	}

	id, _ := app.CreateTransfer(ctx, request)

	result1, _ := app.UpdateTransferStatus(ctx, id, "completed")
	result2, _ := app.UpdateTransferStatus(ctx, id, "completed")

	if result1.Status != result2.Status {
		t.Error("Idempotent calls should return same status")
	}
}
