package db

import (
	"context"
	"ajaib-testing-code/internal/adapters/core/entity"
)

type TransferDBInterface interface {
	CreateTransfer(ctx context.Context, transfer entity.Transfer) (int64, error)
	GetTransferByID(ctx context.Context, id int64) (*entity.Transfer, error)
	GetListTransfer(ctx context.Context) ([]entity.Transfer, error)
	UpdateTransferStatus(ctx context.Context, id int64, status string) error
}
