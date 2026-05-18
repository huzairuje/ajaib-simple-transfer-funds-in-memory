package transfer

import (
	"context"
	"fmt"
	"sync"
	"ajaib-testing-code/internal/adapters/core/entity"
)

type InMemoryRepository struct {
	transfers  []entity.Transfer
	idSequence int64
	mu         sync.RWMutex
}

type Config struct{}

func New(config Config) *InMemoryRepository {
	return &InMemoryRepository{
		transfers:  make([]entity.Transfer, 0),
		idSequence: 1,
	}
}

func (r *InMemoryRepository) CreateTransfer(ctx context.Context, t entity.Transfer) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t.ID = r.idSequence
	r.idSequence++
	r.transfers = append(r.transfers, t)
	return t.ID, nil
}

func (r *InMemoryRepository) GetTransferByID(ctx context.Context, id int64) (*entity.Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.transfers {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("transfer not found")
}

func (r *InMemoryRepository) GetListTransfer(ctx context.Context) ([]entity.Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.transfers, nil
}

func (r *InMemoryRepository) UpdateTransferStatus(ctx context.Context, id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.transfers {
		if r.transfers[i].ID == id {
			r.transfers[i].Status = status
			return nil
		}
	}

	return fmt.Errorf("transfer not found")
}
