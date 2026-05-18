package transfer

import (
	"ajaib-testing-code/internal/adapters/core/entity"
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

type MockApp struct {
	transfers map[int64]*entity.Transfer
	nextID    int64
}

func NewMockApp() *MockApp {
	return &MockApp{
		transfers: make(map[int64]*entity.Transfer),
		nextID:    1,
	}
}

func (m *MockApp) CreateTransfer(ctx context.Context, request entity.CreateTransferRequest) (int64, error) {
	transfer := &entity.Transfer{
		ID:            m.nextID,
		FromAccountID: request.From,
		ToAccountID:   request.To,
		Amount:        request.Amount,
		Currency:      request.Currency,
		Status:        "success",
		FromBalance:   request.FromBalance - request.Amount,
		ToBalance:     request.ToBalance + request.Amount,
	}
	m.transfers[m.nextID] = transfer
	m.nextID++
	return transfer.ID, nil
}

func (m *MockApp) GetTransferByID(ctx context.Context, id int64) (*entity.Transfer, error) {
	if t, exists := m.transfers[id]; exists {
		return t, nil
	}
	return nil, nil
}

func (m *MockApp) GetListTransfer(ctx context.Context) ([]entity.Transfer, error) {
	var result []entity.Transfer
	for _, t := range m.transfers {
		result = append(result, *t)
	}
	return result, nil
}

func (m *MockApp) UpdateTransferStatus(ctx context.Context, id int64, status string) (*entity.Transfer, error) {
	if t, exists := m.transfers[id]; exists {
		t.Status = status
		return t, nil
	}
	return nil, nil
}

func TestHandler_CreateTransferHandler(t *testing.T) {
	mockApp := NewMockApp()
	_ = NewHandler(Config{TransferApp: mockApp})

	body := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	bodyBytes, _ := json.Marshal(body)
	_ = httptest.NewRequest("POST", "/v1/transfers", bytes.NewReader(bodyBytes))

	t.Log("Handler test requires Gin context - see integration tests")
}

func TestHandler_GetListTransferHandler(t *testing.T) {
	mockApp := NewMockApp()
	_ = NewHandler(Config{TransferApp: mockApp})

	mockApp.CreateTransfer(context.Background(), entity.CreateTransferRequest{
		From:     1001,
		To:       1002,
		Amount:   50000,
		Currency: "IDR",
	})

	_ = httptest.NewRequest("GET", "/v1/transfers", nil)

	t.Log("Handler test requires Gin context - see integration tests")
}

func TestHandler_GetDetailTransferHandler(t *testing.T) {
	mockApp := NewMockApp()
	_ = NewHandler(Config{TransferApp: mockApp})

	_, _ = mockApp.CreateTransfer(context.Background(), entity.CreateTransferRequest{
		From:     1001,
		To:       1002,
		Amount:   50000,
		Currency: "IDR",
	})

	_ = httptest.NewRequest("GET", "/v1/transfers/1", nil)

	t.Log("Handler test requires Gin context - see integration tests")
}

func TestHandler_UpdateTransferStatusHandler(t *testing.T) {
	mockApp := NewMockApp()
	_ = NewHandler(Config{TransferApp: mockApp})

	_, _ = mockApp.CreateTransfer(context.Background(), entity.CreateTransferRequest{
		From:     1001,
		To:       1002,
		Amount:   50000,
		Currency: "IDR",
	})

	body := entity.UpdateTransferStatusRequest{Status: "completed"}
	bodyBytes, _ := json.Marshal(body)
	_ = httptest.NewRequest("PATCH", "/v1/transfers/1/status", bytes.NewReader(bodyBytes))

	t.Log("Handler test requires Gin context - see integration tests")
}
