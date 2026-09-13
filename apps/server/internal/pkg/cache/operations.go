package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get fetches JSON data for key and unmarshals it into dest.
// Returns (true, nil) on a cache hit, (false, nil) on a cache miss or if Redis is unavailable.
func (c *Cache) Get(ctx context.Context, key string, dest any) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		slog.Error("cache: get error", "key", key, "error", err)
		return false, nil
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		slog.Error("cache: unmarshal error", "key", key, "error", err)
		return false, nil
	}

	return true, nil
}

// Set serializes val to JSON and stores it in Redis with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}

	data, err := json.Marshal(val)
	if err != nil {
		slog.Error("cache: marshal error", "key", key, "error", err)
		return err
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		slog.Error("cache: set error", "key", key, "error", err)
		return nil
	}

	return nil
}

// Delete removes one or more keys from Redis.
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if c == nil || c.client == nil || len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		slog.Error("cache: delete error", "error", err)
		return err
	}

	return nil
}

// Remember implements the standard cache-aside pattern:
// 1. Returns cached value if present.
// 2. Otherwise calls fetch(), stores result in cache for ttl, and returns it.
func Remember[T any](ctx context.Context, c *Cache, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	if c != nil && c.client != nil {
		var cached T
		hit, err := c.Get(ctx, key, &cached)
		if err == nil && hit {
			return cached, nil
		}
	}

	data, err := fetch()
	if err != nil {
		var zero T
		return zero, err
	}

	if c != nil && c.client != nil {
		_ = c.Set(ctx, key, data, ttl)
	}

	return data, nil
}
