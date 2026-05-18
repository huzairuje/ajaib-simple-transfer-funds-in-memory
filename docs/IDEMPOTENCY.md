# Idempotency Implementation Guide

## Overview

Idempotency ensures that duplicate requests produce the same result without side effects. This is critical for financial transactions and status updates where network failures or retries could cause duplicate operations.

## Current Implementation

### In-Memory Storage

The current implementation uses an in-memory map to store idempotency records:

```go
type IdempotencyRecord struct {
    Key        string
    TransferID int64
    Status     string
}

type InMemoryRepository struct {
    idempotencyKeys map[string]IdempotencyRecord
    mu              sync.RWMutex
}
```

### How It Works

1. **Key Generation**: Internally generated as `transfer:{id}:status:{status}`
2. **Duplicate Detection**: Checks if key exists in memory
3. **Idempotent Response**: Returns cached result for duplicate requests
4. **Conflict Detection**: Returns error if same key with different parameters

### Usage

```bash
curl -X PATCH http://localhost:3400/v1/transfers/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":"completed"}'
```

Duplicate requests with same transfer ID and status return the same result.

## Scaling to Production

### Phase 1: Redis (Recommended for most cases)

Redis provides distributed, fast idempotency key storage:

```go
type RedisRepository struct {
    client *redis.Client
    mu     sync.RWMutex
}

func (r *RedisRepository) UpdateTransferStatus(id int64, status string, key string) (*Transfer, error) {
    // Check if key exists in Redis
    result, err := r.client.Get(ctx, key).Result()
    if err == nil {
        // Key exists, return cached result
        return parseTransfer(result), nil
    }
    
    // Update transfer
    transfer := updateTransfer(id, status)
    
    // Store in Redis with TTL (24 hours)
    r.client.Set(ctx, key, marshal(transfer), 24*time.Hour)
    
    return transfer, nil
}
```

**Advantages:**
- Fast in-memory access
- Distributed across multiple instances
- Built-in TTL for automatic cleanup
- Atomic operations

**Setup:**
```bash
docker run -d -p 6379:6379 redis:latest
```

### Phase 2: PostgreSQL (For persistence)

Store idempotency records in database for long-term audit trails:

```go
type PostgresRepository struct {
    db *sql.DB
}

func (r *PostgresRepository) UpdateTransferStatus(id int64, status string, key string) (*Transfer, error) {
    // Check existing idempotency record
    var existingTransferID int64
    err := r.db.QueryRow(
        "SELECT transfer_id FROM idempotency_keys WHERE key = $1",
        key,
    ).Scan(&existingTransferID)
    
    if err == nil && existingTransferID == id {
        // Return existing transfer
        return r.GetTransferByID(id)
    }
    
    // Update transfer in transaction
    tx, _ := r.db.Begin()
    defer tx.Rollback()
    
    // Update transfer
    tx.Exec("UPDATE transfers SET status = $1 WHERE id = $2", status, id)
    
    // Store idempotency key
    tx.Exec(
        "INSERT INTO idempotency_keys (key, transfer_id, status, created_at) VALUES ($1, $2, $3, NOW())",
        key, id, status,
    )
    
    tx.Commit()
    return r.GetTransferByID(id)
}
```

**Schema:**
```sql
CREATE TABLE idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    transfer_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_key (key),
    INDEX idx_transfer_id (transfer_id)
);
```

### Phase 3: Hybrid Approach (Recommended for scale)

Combine Redis (fast) + PostgreSQL (persistent):

```go
type HybridRepository struct {
    redis    *redis.Client
    postgres *sql.DB
}

func (r *HybridRepository) UpdateTransferStatus(id int64, status string, key string) (*Transfer, error) {
    // Check Redis first (fast path)
    if cached, err := r.redis.Get(ctx, key).Result(); err == nil {
        return parseTransfer(cached), nil
    }
    
    // Check PostgreSQL (fallback)
    if existing, err := r.checkPostgres(key, id); err == nil {
        // Repopulate Redis
        r.redis.Set(ctx, key, marshal(existing), 24*time.Hour)
        return existing, nil
    }
    
    // Update transfer
    transfer := updateTransfer(id, status)
    
    // Store in both Redis and PostgreSQL
    r.redis.Set(ctx, key, marshal(transfer), 24*time.Hour)
    r.storeInPostgres(key, id, status)
    
    return transfer, nil
}
```

## Architecture Considerations

### TTL (Time To Live)

- **In-Memory**: No cleanup (memory leak risk)
- **Redis**: 24-48 hours recommended
- **PostgreSQL**: Archive old records, keep recent ones

### Consistency

- **Strong Consistency**: PostgreSQL + Redis with sync writes
- **Eventual Consistency**: Redis with async PostgreSQL backup

### Failure Scenarios

| Scenario | In-Memory | Redis | PostgreSQL | Hybrid |
|----------|-----------|-------|------------|--------|
| Server restart | ❌ Lost | ✅ Persists | ✅ Persists | ✅ Persists |
| Network partition | ✅ Works | ❌ Fails | ✅ Works | ✅ Works |
| High throughput | ⚠️ Limited | ✅ Excellent | ⚠️ Limited | ✅ Excellent |
| Cost | ✅ Free | ⚠️ Moderate | ⚠️ Moderate | ⚠️ Higher |

## Migration Path

1. **Start**: In-memory (development/testing)
2. **Scale**: Add Redis for distributed systems
3. **Mature**: Add PostgreSQL for audit trail
4. **Production**: Hybrid approach with monitoring

## Best Practices

1. **Key Format**: Use deterministic, collision-free keys
2. **TTL**: Set appropriate expiration (24-48 hours typical)
3. **Monitoring**: Track idempotency hit rates
4. **Logging**: Log all idempotency key operations
5. **Testing**: Test duplicate request scenarios
6. **Documentation**: Document key generation strategy

## Example: Switching to Redis

Update `main.go`:

```go
import "github.com/redis/go-redis/v9"

func main() {
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    repoService := service.NewRedisRepository(redisClient)
    handler := handler.NewHandler(repoService)
    
    // ... rest of setup
}
```

## Monitoring

Track these metrics:

- Idempotency key hit rate
- Cache hit/miss ratio
- Duplicate request frequency
- Key storage size
- TTL expiration rate

## References

- [RFC 7231 - HTTP Semantics (Idempotent Methods)](https://tools.ietf.org/html/rfc7231#section-4.2.2)
- [Stripe Idempotency Documentation](https://stripe.com/docs/api/idempotent_requests)
- [Redis Documentation](https://redis.io/docs/)
