package e2e

import (
	transferApp "ajaib-testing-code/internal/adapters/app/transfer"
	"ajaib-testing-code/internal/adapters/core/entity"
	transferCore "ajaib-testing-code/internal/adapters/core/transfer"
	transferHandler "ajaib-testing-code/internal/adapters/framework/primary/rest_fiber/transfer"
	idempotencyCache "ajaib-testing-code/internal/adapters/framework/secondary/repository/cache/idempotency"
	transferDB "ajaib-testing-code/internal/adapters/framework/secondary/repository/db/transfer"
	"ajaib-testing-code/router"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:3401"

func setupServer() *router.Router {
	dbRepo := transferDB.New(transferDB.Config{})
	cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

	core := transferCore.New(transferCore.Config{
		DB:    dbRepo,
		Cache: cacheRepo,
	})

	app := transferApp.New(transferApp.Config{
		Core: core,
	})

	handler := transferHandler.NewHandler(transferHandler.Config{
		TransferApp: app,
	})

	return router.NewRouter(router.Config{
		TransferHandler: handler,
	})
}

func TestE2E_CreateTransfer(t *testing.T) {
	r := setupServer()
	go r.Run("3401")
	time.Sleep(500 * time.Millisecond)

	requestBody := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	body, _ := json.Marshal(requestBody)
	resp, err := http.Post(baseURL+"/v1/transfers", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response entity.TransferResponse
	json.NewDecoder(resp.Body).Decode(&response)

	if response.ID == 0 {
		t.Error("Expected non-zero transfer ID")
	}
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}
}

func TestE2E_GetTransferByID(t *testing.T) {
	r := setupServer()
	go r.Run("3402")
	time.Sleep(500 * time.Millisecond)

	requestBody := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	body, _ := json.Marshal(requestBody)
	createResp, _ := http.Post("http://localhost:3402/v1/transfers", "application/json", bytes.NewReader(body))
	var createResponse entity.TransferResponse
	json.NewDecoder(createResp.Body).Decode(&createResponse)
	createResp.Body.Close()

	resp, err := http.Get(fmt.Sprintf("http://localhost:3402/v1/transfers/%d", createResponse.ID))
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var transfer entity.Transfer
	json.NewDecoder(resp.Body).Decode(&transfer)

	if transfer.ID != createResponse.ID {
		t.Errorf("Expected ID %d, got %d", createResponse.ID, transfer.ID)
	}
	if transfer.Amount != requestBody.Amount {
		t.Errorf("Expected Amount %d, got %d", requestBody.Amount, transfer.Amount)
	}
}

func TestE2E_GetListTransfers(t *testing.T) {
	r := setupServer()
	go r.Run("3403")
	time.Sleep(500 * time.Millisecond)

	requests := []entity.CreateTransferRequest{
		{From: 1001, To: 1002, Amount: 50000, Currency: "IDR", FromBalance: 100000, ToBalance: 50000},
		{From: 1003, To: 1004, Amount: 75000, Currency: "IDR", FromBalance: 150000, ToBalance: 75000},
	}

	for _, req := range requests {
		body, _ := json.Marshal(req)
		http.Post("http://localhost:3403/v1/transfers", "application/json", bytes.NewReader(body))
	}

	resp, err := http.Get("http://localhost:3403/v1/transfers")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var transfers []entity.Transfer
	json.NewDecoder(resp.Body).Decode(&transfers)

	if len(transfers) != 2 {
		t.Errorf("Expected 2 transfers, got %d", len(transfers))
	}
}

func TestE2E_UpdateTransferStatus(t *testing.T) {
	r := setupServer()
	go r.Run("3404")
	time.Sleep(500 * time.Millisecond)

	createBody := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	body, _ := json.Marshal(createBody)
	createResp, _ := http.Post("http://localhost:3404/v1/transfers", "application/json", bytes.NewReader(body))
	var createResponse entity.TransferResponse
	json.NewDecoder(createResp.Body).Decode(&createResponse)
	createResp.Body.Close()

	updateBody := entity.UpdateTransferStatusRequest{Status: "completed"}
	updateBodyBytes, _ := json.Marshal(updateBody)

	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("http://localhost:3404/v1/transfers/%d/status", createResponse.ID), bytes.NewReader(updateBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PATCH request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var transfer entity.Transfer
	json.NewDecoder(resp.Body).Decode(&transfer)

	if transfer.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", transfer.Status)
	}
}

func TestE2E_UpdateTransferStatus_Idempotent(t *testing.T) {
	r := setupServer()
	go r.Run("3405")
	time.Sleep(500 * time.Millisecond)

	createBody := entity.CreateTransferRequest{
		From:        1001,
		To:          1002,
		Amount:      50000,
		Currency:    "IDR",
		FromBalance: 100000,
		ToBalance:   50000,
	}

	body, _ := json.Marshal(createBody)
	createResp, _ := http.Post("http://localhost:3405/v1/transfers", "application/json", bytes.NewReader(body))
	var createResponse entity.TransferResponse
	json.NewDecoder(createResp.Body).Decode(&createResponse)
	createResp.Body.Close()

	updateBody := entity.UpdateTransferStatusRequest{Status: "completed"}
	updateBodyBytes, _ := json.Marshal(updateBody)

	client := &http.Client{}

	req1, _ := http.NewRequest("PATCH", fmt.Sprintf("http://localhost:3405/v1/transfers/%d/status", createResponse.ID), bytes.NewReader(updateBodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	resp1, _ := client.Do(req1)
	var transfer1 entity.Transfer
	json.NewDecoder(resp1.Body).Decode(&transfer1)
	resp1.Body.Close()

	req2, _ := http.NewRequest("PATCH", fmt.Sprintf("http://localhost:3405/v1/transfers/%d/status", createResponse.ID), bytes.NewReader(updateBodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := client.Do(req2)
	var transfer2 entity.Transfer
	json.NewDecoder(resp2.Body).Decode(&transfer2)
	resp2.Body.Close()

	if transfer1.Status != transfer2.Status {
		t.Error("Idempotent requests should return same status")
	}
	if transfer1.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", transfer1.Status)
	}
}

func TestE2E_InvalidRequest(t *testing.T) {
	r := setupServer()
	go r.Run("3406")
	time.Sleep(500 * time.Millisecond)

	invalidBody := []byte(`{"invalid": "data"}`)
	resp, _ := http.Post("http://localhost:3406/v1/transfers", "application/json", bytes.NewReader(invalidBody))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}
