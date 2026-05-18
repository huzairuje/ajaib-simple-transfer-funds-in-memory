package transfer

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"ajaib-testing-code/internal/adapters/core/entity"
	"ajaib-testing-code/internal/ports/app"
)

type Handler struct {
	transferApp app.TransferInterface
}

type Config struct {
	TransferApp app.TransferInterface
}

func NewHandler(config Config) *Handler {
	return &Handler{
		transferApp: config.TransferApp,
	}
}

func (h *Handler) CreateTransferHandler(c *gin.Context) {
	startNow := time.Now()
	funcLog := "CreateTransferHandler"
	defer func() {
		duration := time.Since(startNow)
		slog.InfoContext(c, fmt.Sprintf("%s executed", funcLog), "duration", duration)
	}()

	var req entity.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to bind request", funcLog), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transferID, err := h.transferApp.CreateTransfer(c, req)
	if err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to create transfer", funcLog), "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := entity.TransferResponse{
		ID:          transferID,
		Status:      "success",
		FromBalance: req.FromBalance - req.Amount,
		ToBalance:   req.ToBalance + req.Amount,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetDetailTransferHandler(c *gin.Context) {
	startNow := time.Now()
	funcLog := "GetDetailTransferHandler"
	defer func() {
		duration := time.Since(startNow)
		slog.InfoContext(c, fmt.Sprintf("%s executed", funcLog), "duration", duration)
	}()

	idStr := c.Param("id")
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to convert id to int", funcLog), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	transfer, err := h.transferApp.GetTransferByID(c, int64(idInt))
	if err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to get transfer", funcLog), "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "transfer not found"})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

func (h *Handler) GetListTransferHandler(c *gin.Context) {
	startNow := time.Now()
	funcLog := "GetListTransferHandler"
	defer func() {
		duration := time.Since(startNow)
		slog.InfoContext(c, fmt.Sprintf("%s executed", funcLog), "duration", duration)
	}()

	transfers, err := h.transferApp.GetListTransfer(c)
	if err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to get transfers", funcLog), "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfers)
}

func (h *Handler) UpdateTransferStatusHandler(c *gin.Context) {
	startNow := time.Now()
	funcLog := "UpdateTransferStatusHandler"
	defer func() {
		duration := time.Since(startNow)
		slog.InfoContext(c, fmt.Sprintf("%s executed", funcLog), "duration", duration)
	}()

	idStr := c.Param("id")
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to convert id to int", funcLog), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req entity.UpdateTransferStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to bind request", funcLog), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.transferApp.UpdateTransferStatus(c, int64(idInt), req.Status)
	if err != nil {
		slog.ErrorContext(c, fmt.Sprintf("%s failed to update status", funcLog), "error", err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfer)
}
