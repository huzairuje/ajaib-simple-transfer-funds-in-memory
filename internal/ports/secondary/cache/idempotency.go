package cache

import (
	"context"
	"ajaib-testing-code/internal/adapters/core/entity"
)

type IdempotencyCacheInterface interface {
	Get(ctx context.Context, key string) (*entity.IdempotencyRecord, error)
	Set(ctx context.Context, key string, record entity.IdempotencyRecord) error
	Exists(ctx context.Context, key string) bool
}
