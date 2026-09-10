package cache

import (
    "context"
    "encoding/json"
    "log/slog"
    "time"

    "github.com/redis/go-redis/v9"
)

// Get fetches data from Redis and unmarshals it into dest.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
    if c == nil || c.client == nil {
        return false, nil
    }
    val, err := c.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return false, nil
    } else if err != nil {
        slog.Error("cache get error", "key", key, "error", err)
        return false, nil
    }
    if err := json.Unmarshal([]byte(val), dest); err != nil {
        slog.Error("cache json unmarshal error", "key", key, "error", err)
        return false, nil
    }
    return true, nil
}

// GetGeneric fetches data from Redis into a pointer of generic type T.
func Get[T interface{}](c *Cache, ctx context.Context, key string, dest *T) (bool, error) {
    return c.Get(ctx, key, dest)
}

// Set marshals value into JSON and stores it in Redis with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
    if c == nil || c.client == nil {
        return nil
    }
    data, err := json.Marshal(val)
    if err != nil {
        slog.Error("cache json marshal error", "key", key, "error", err)
        return err
    }
    if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
        slog.Error("cache set error", "key", key, "error", err)
        return nil
    }
    return nil
}

// SetGeneric stores a generic value of type T in Redis with given TTL.
func Set[T interface{}](c *Cache, ctx context.Context, key string, val T, ttl time.Duration) error {
    return c.Set(ctx, key, val, ttl)
}

// Fetch returns the cached value at key if present; otherwise it calls fn,
// caches the result for ttl, and returns it — the read-check-set pattern
// otherwise hand-written at the top of most cached service methods.
func Fetch[T interface{}](ctx context.Context, c *Cache, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
    var cached T
    if hit, _ := c.Get(ctx, key, &cached); hit {
        return cached, nil
    }
    result, err := fn()
    if err != nil {
        var zero T
        return zero, err
    }
    _ = c.Set(ctx, key, result, ttl)
    return result, nil
}

// Delete removes specific keys from Redis.
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
    if c == nil || c.client == nil || len(keys) == 0 {
        return nil
    }
    if err := c.client.Del(ctx, keys...).Err(); err != nil {
        slog.Error("cache delete error", "error", err)
    }
    return nil
}

// DeleteByPattern scans for matching keys and deletes them.
func (c *Cache) DeleteByPattern(ctx context.Context, pattern string) error {
    if c == nil || c.client == nil {
        return nil
    }
    var cursor uint64
    var keys []string
    for {
        var err error
        var k []string
        k, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
            slog.Error("cache delete-by-pattern scan error", "pattern", pattern, "error", err)
            return nil
        }
        keys = append(keys, k...)
        if cursor == 0 {
            break
        }
    }
    if len(keys) > 0 {
        if err := c.client.Del(ctx, keys...).Err(); err != nil {
            slog.Error("cache delete-by-pattern delete error", "pattern", pattern, "error", err)
        }
    }
    return nil
}
