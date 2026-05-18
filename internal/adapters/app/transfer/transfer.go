package transfer

import (
	"context"
	"fmt"
	"ajaib-testing-code/internal/adapters/core/entity"
	"ajaib-testing-code/internal/ports/core"
)

type transfer struct {
	core core.TransferInterface
}

type Config struct {
	Core core.TransferInterface
}

func New(config Config) *transfer {
	return &transfer{
		core: config.Core,
	}
}

func (t *transfer) CreateTransfer(ctx context.Context, request entity.CreateTransferRequest) (int64, error) {
	transfer := entity.Transfer{
		FromAccountID: request.From,
		ToAccountID:   request.To,
		Amount:        request.Amount,
		Currency:      request.Currency,
		FromBalance:   request.FromBalance - request.Amount,
		ToBalance:     request.ToBalance + request.Amount,
		Status:        "success",
	}

	return t.core.CreateTransfer(ctx, transfer)
}

func (t *transfer) GetTransferByID(ctx context.Context, id int64) (*entity.Transfer, error) {
	return t.core.GetTransferByID(ctx, id)
}

func (t *transfer) GetListTransfer(ctx context.Context) ([]entity.Transfer, error) {
	return t.core.GetListTransfer(ctx)
}

func (t *transfer) UpdateTransferStatus(ctx context.Context, id int64, status string) (*entity.Transfer, error) {
	idempotencyKey := fmt.Sprintf("transfer:%d:status:%s", id, status)
	return t.core.UpdateTransferStatus(ctx, id, status, idempotencyKey)
}
