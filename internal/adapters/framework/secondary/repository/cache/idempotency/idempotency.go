package idempotency

import (
	"context"
	"fmt"
	"sync"
	"ajaib-testing-code/internal/adapters/core/entity"
)

type InMemoryCache struct {
	records map[string]entity.IdempotencyRecord
	mu      sync.RWMutex
}

type Config struct{}

func New(config Config) *InMemoryCache {
	return &InMemoryCache{
		records: make(map[string]entity.IdempotencyRecord),
	}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) (*entity.IdempotencyRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if record, exists := c.records[key]; exists {
		return &record, nil
	}
	return nil, fmt.Errorf("key not found")
}

func (c *InMemoryCache) Set(ctx context.Context, key string, record entity.IdempotencyRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.records[key] = record
	return nil
}

func (c *InMemoryCache) Exists(ctx context.Context, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.records[key]
	return exists
}
