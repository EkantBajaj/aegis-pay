package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

// NewClient initializes the connection to the Redis container
func NewClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Client{rdb: rdb}
}

// CheckAndLock attempts to set a key only if it doesn't exist.
// Returns true if the lock was acquired, false if it already exists.
func (c *Client) CheckAndLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// SetNX = "SET if Not eXists"
	// We mark it as IN_PROGRESS initially
	ok, err := c.rdb.SetNX(ctx, "idempotency:"+key, "IN_PROGRESS", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis error: %w", err)
	}
	return ok, nil
}

// SetResult saves the final response of the transaction so it can be replayed
func (c *Client) SetResult(ctx context.Context, key string, result string, ttl time.Duration) error {
	return c.rdb.Set(ctx, "idempotency:"+key, result, ttl).Err()
}

// GetResult retrieves the cached response if it exists
func (c *Client) GetResult(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, "idempotency:"+key).Result()
}
