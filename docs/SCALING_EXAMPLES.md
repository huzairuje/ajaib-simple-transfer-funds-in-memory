# Scaling Implementation Examples

## Redis Implementation

### Installation

```bash
go get github.com/redis/go-redis/v9
```

### Redis Repository

Create `service/redis_repository.go`:

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client    *redis.Client
	transfers []Transfer
	mu        sync.RWMutex
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client:    client,
		transfers: make([]Transfer, 0),
	}
}

func (r *RedisRepository) CreateTransfer(t Transfer) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t.ID = int64(len(r.transfers)) + 1
	r.transfers = append(r.transfers, t)
	return t.ID, nil
}

func (r *RedisRepository) GetTransferByID(id int64) (*Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.transfers {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("transfer not found")
}

func (r *RedisRepository) GetListTransfer() ([]Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.transfers, nil
}

func (r *RedisRepository) UpdateTransferStatus(id int64, status string, idempotencyKey string) (*Transfer, error) {
	ctx := context.Background()

	// Check Redis for existing idempotency key
	cachedResult, err := r.client.Get(ctx, idempotencyKey).Result()
	if err == nil {
		var record IdempotencyRecord
		json.Unmarshal([]byte(cachedResult), &record)
		if record.TransferID == id && record.Status == status {
			transfer, _ := r.GetTransferByID(id)
			return transfer, nil
		}
		return nil, fmt.Errorf("idempotency key conflict: different operation")
	}

	// Update transfer
	r.mu.Lock()
	for i := range r.transfers {
		if r.transfers[i].ID == id {
			r.transfers[i].Status = status

			// Store in Redis with 24-hour TTL
			record := IdempotencyRecord{
				Key:        idempotencyKey,
				TransferID: id,
				Status:     status,
			}
			data, _ := json.Marshal(record)
			r.client.Set(ctx, idempotencyKey, string(data), 24*time.Hour)

			r.mu.Unlock()
			return &r.transfers[i], nil
		}
	}
	r.mu.Unlock()

	return nil, fmt.Errorf("transfer not found")
}
```

### Usage in main.go

```go
import "github.com/redis/go-redis/v9"

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	repoService := service.NewRedisRepository(redisClient)
	handler := handler.NewHandler(repoService)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.POST("/v1/transfers", handler.CreateTransferHandler)
	r.GET("/v1/transfers/:id", handler.GetDetailTransferHandler)
	r.GET("/v1/transfers", handler.GetListTransferHandler)
	r.PATCH("/v1/transfers/:id/status", handler.UpdateTransferStatusHandler)

	_ = r.Run(":3400")
}
```

## PostgreSQL Implementation

### Installation

```bash
go get github.com/lib/pq
```

### Database Schema

```sql
CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT NOT NULL,
    to_account_id BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(50) NOT NULL,
    from_balance BIGINT NOT NULL,
    to_balance BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    transfer_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (transfer_id) REFERENCES transfers(id),
    INDEX idx_key (key),
    INDEX idx_transfer_id (transfer_id)
);
```

### PostgreSQL Repository

Create `service/postgres_repository.go`:

```go
package service

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db        *sql.DB
	transfers []Transfer
	mu        sync.RWMutex
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db:        db,
		transfers: make([]Transfer, 0),
	}
}

func (r *PostgresRepository) CreateTransfer(t Transfer) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var id int64
	err := r.db.QueryRow(
		"INSERT INTO transfers (from_account_id, to_account_id, amount, currency, status, from_balance, to_balance) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		t.FromAccountID, t.ToAccountID, t.Amount, t.Currency, t.Status, t.FromBalance, t.ToBalance,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	t.ID = id
	r.transfers = append(r.transfers, t)
	return id, nil
}

func (r *PostgresRepository) GetTransferByID(id int64) (*Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var t Transfer
	err := r.db.QueryRow(
		"SELECT id, from_account_id, to_account_id, amount, currency, status, from_balance, to_balance FROM transfers WHERE id = $1",
		id,
	).Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Currency, &t.Status, &t.FromBalance, &t.ToBalance)

	if err != nil {
		return nil, fmt.Errorf("transfer not found")
	}

	return &t, nil
}

func (r *PostgresRepository) GetListTransfer() ([]Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query("SELECT id, from_account_id, to_account_id, amount, currency, status, from_balance, to_balance FROM transfers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []Transfer
	for rows.Next() {
		var t Transfer
		rows.Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Currency, &t.Status, &t.FromBalance, &t.ToBalance)
		transfers = append(transfers, t)
	}

	return transfers, nil
}

func (r *PostgresRepository) UpdateTransferStatus(id int64, status string, idempotencyKey string) (*Transfer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if idempotency key exists
	var existingTransferID int64
	err := r.db.QueryRow(
		"SELECT transfer_id FROM idempotency_keys WHERE key = $1",
		idempotencyKey,
	).Scan(&existingTransferID)

	if err == nil {
		if existingTransferID == id {
			return r.GetTransferByID(id)
		}
		return nil, fmt.Errorf("idempotency key conflict: different operation")
	}

	// Update transfer
	_, err = r.db.Exec("UPDATE transfers SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return nil, err
	}

	// Store idempotency key
	_, err = r.db.Exec(
		"INSERT INTO idempotency_keys (key, transfer_id, status) VALUES ($1, $2, $3)",
		idempotencyKey, id, status,
	)

	if err != nil {
		return nil, err
	}

	return r.GetTransferByID(id)
}
```

## Hybrid Approach (Redis + PostgreSQL)

Create `service/hybrid_repository.go`:

```go
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type HybridRepository struct {
	redis     *redis.Client
	db        *sql.DB
	transfers []Transfer
	mu        sync.RWMutex
}

func NewHybridRepository(redis *redis.Client, db *sql.DB) *HybridRepository {
	return &HybridRepository{
		redis:     redis,
		db:        db,
		transfers: make([]Transfer, 0),
	}
}

func (r *HybridRepository) CreateTransfer(t Transfer) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var id int64
	err := r.db.QueryRow(
		"INSERT INTO transfers (from_account_id, to_account_id, amount, currency, status, from_balance, to_balance) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		t.FromAccountID, t.ToAccountID, t.Amount, t.Currency, t.Status, t.FromBalance, t.ToBalance,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	t.ID = id
	r.transfers = append(r.transfers, t)
	return id, nil
}

func (r *HybridRepository) GetTransferByID(id int64) (*Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var t Transfer
	err := r.db.QueryRow(
		"SELECT id, from_account_id, to_account_id, amount, currency, status, from_balance, to_balance FROM transfers WHERE id = $1",
		id,
	).Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Currency, &t.Status, &t.FromBalance, &t.ToBalance)

	if err != nil {
		return nil, fmt.Errorf("transfer not found")
	}

	return &t, nil
}

func (r *HybridRepository) GetListTransfer() ([]Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query("SELECT id, from_account_id, to_account_id, amount, currency, status, from_balance, to_balance FROM transfers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []Transfer
	for rows.Next() {
		var t Transfer
		rows.Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Currency, &t.Status, &t.FromBalance, &t.ToBalance)
		transfers = append(transfers, t)
	}

	return transfers, nil
}

func (r *HybridRepository) UpdateTransferStatus(id int64, status string, idempotencyKey string) (*Transfer, error) {
	ctx := context.Background()
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check Redis first (fast path)
	cachedResult, err := r.redis.Get(ctx, idempotencyKey).Result()
	if err == nil {
		var record IdempotencyRecord
		json.Unmarshal([]byte(cachedResult), &record)
		if record.TransferID == id && record.Status == status {
			transfer, _ := r.GetTransferByID(id)
			return transfer, nil
		}
		return nil, fmt.Errorf("idempotency key conflict: different operation")
	}

	// Check PostgreSQL (fallback)
	var existingTransferID int64
	err = r.db.QueryRow(
		"SELECT transfer_id FROM idempotency_keys WHERE key = $1",
		idempotencyKey,
	).Scan(&existingTransferID)

	if err == nil {
		if existingTransferID == id {
			transfer, _ := r.GetTransferByID(id)
			// Repopulate Redis
			record := IdempotencyRecord{
				Key:        idempotencyKey,
				TransferID: id,
				Status:     status,
			}
			data, _ := json.Marshal(record)
			r.redis.Set(ctx, idempotencyKey, string(data), 24*time.Hour)
			return transfer, nil
		}
		return nil, fmt.Errorf("idempotency key conflict: different operation")
	}

	// Update transfer
	_, err = r.db.Exec("UPDATE transfers SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return nil, err
	}

	// Store in PostgreSQL
	_, err = r.db.Exec(
		"INSERT INTO idempotency_keys (key, transfer_id, status) VALUES ($1, $2, $3)",
		idempotencyKey, id, status,
	)

	if err != nil {
		return nil, err
	}

	// Store in Redis with 24-hour TTL
	record := IdempotencyRecord{
		Key:        idempotencyKey,
		TransferID: id,
		Status:     status,
	}
	data, _ := json.Marshal(record)
	r.redis.Set(ctx, idempotencyKey, string(data), 24*time.Hour)

	return r.GetTransferByID(id)
}
```

## Docker Compose Setup

Create `docker-compose.yml` for local development:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
      POSTGRES_DB: transfers
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

Start services:
```bash
docker-compose up -d
```

## Performance Comparison

| Operation | In-Memory | Redis | PostgreSQL | Hybrid |
|-----------|-----------|-------|------------|--------|
| Create Transfer | <1ms | 2-5ms | 5-10ms | 5-10ms |
| Get Transfer | <1ms | 2-5ms | 5-10ms | <1ms (cached) |
| Update Status | <1ms | 2-5ms | 5-10ms | 2-5ms |
| Idempotency Check | <1ms | 2-5ms | 5-10ms | <1ms (Redis hit) |

## Monitoring & Observability

Add metrics collection:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
	idempotencyHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "idempotency_hits_total",
		Help: "Total idempotency key hits",
	})
	idempotencyMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "idempotency_misses_total",
		Help: "Total idempotency key misses",
	})
)

func (r *HybridRepository) UpdateTransferStatus(id int64, status string, key string) (*Transfer, error) {
	// Check Redis
	if _, err := r.redis.Get(ctx, key).Result(); err == nil {
		idempotencyHits.Inc()
		// ... return cached
	}
	idempotencyMisses.Inc()
	// ... process
}
```

## Deployment Checklist

- [ ] Set up Redis cluster for high availability
- [ ] Configure PostgreSQL replication
- [ ] Enable connection pooling (PgBouncer)
- [ ] Set up monitoring and alerting
- [ ] Configure backup strategy
- [ ] Load test with expected traffic
- [ ] Document rollback procedures
- [ ] Set up graceful shutdown handlers
