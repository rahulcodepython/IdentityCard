package cache

import (
    "context"
    "fmt"

    "github.com/redis/go-redis/v9"
)

// Cache wraps the Redis client and exposes caching, locking, and invalidation operations.
type Cache struct {
    client *redis.Client
}

// NewCache initializes a new Cache instance with the provided Redis client.
func NewCache(client *redis.Client) *Cache {
    return &Cache{client: client}
}

// New is an alias for NewCache for backward compatibility.
func New(client *redis.Client) *Cache {
    return NewCache(client)
}

// Ping checks the health status of the Redis connection.
func (c *Cache) Ping(ctx context.Context) error {
    if c == nil || c.client == nil {
        return fmt.Errorf("redis cache client not initialized")
    }
    return c.client.Ping(ctx).Err()
}
