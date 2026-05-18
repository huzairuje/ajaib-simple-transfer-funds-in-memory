package transfer

import (
	"context"
	"fmt"
	"sync"
	"ajaib-testing-code/internal/adapters/core/entity"
	"ajaib-testing-code/internal/ports/secondary/cache"
	"ajaib-testing-code/internal/ports/secondary/db"
)

type transfer struct {
	db    db.TransferDBInterface
	cache cache.IdempotencyCacheInterface
	mu    sync.RWMutex
}

type Config struct {
	DB    db.TransferDBInterface
	Cache cache.IdempotencyCacheInterface
}

func New(config Config) *transfer {
	return &transfer{
		db:    config.DB,
		cache: config.Cache,
	}
}

func (t *transfer) CreateTransfer(ctx context.Context, transfer entity.Transfer) (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.db.CreateTransfer(ctx, transfer)
}

func (t *transfer) GetTransferByID(ctx context.Context, id int64) (*entity.Transfer, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.db.GetTransferByID(ctx, id)
}

func (t *transfer) GetListTransfer(ctx context.Context) ([]entity.Transfer, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.db.GetListTransfer(ctx)
}

func (t *transfer) UpdateTransferStatus(ctx context.Context, id int64, status string, idempotencyKey string) (*entity.Transfer, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if idempotencyKey != "" {
		if record, err := t.cache.Get(ctx, idempotencyKey); err == nil {
			if record.TransferID == id && record.Status == status {
				return t.db.GetTransferByID(ctx, id)
			}
			return nil, fmt.Errorf("idempotency key conflict: different operation")
		}
	}

	err := t.db.UpdateTransferStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	if idempotencyKey != "" {
		t.cache.Set(ctx, idempotencyKey, entity.IdempotencyRecord{
			Key:        idempotencyKey,
			TransferID: id,
			Status:     status,
		})
	}

	return t.db.GetTransferByID(ctx, id)
}
