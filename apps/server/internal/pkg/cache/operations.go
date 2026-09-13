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

// Entry represents the result of a cache lookup or fallback computation.
type Entry[T any] struct {
    Data     T
    Key      string
    CacheHit bool
}

// FetchOptions configures cache lookup, fallback computation, and hit/miss lifecycle hooks.
type FetchOptions[T any] struct {
    Key     string
    TTL     time.Duration
    Fetch   func(ctx context.Context) (T, error)
    IsValid func(val T) bool
    OnHit   func(ctx context.Context, key string, val T)
    OnMiss  func(ctx context.Context, key string, val T)
}

// FetchOrCompute retrieves the value for the given key from Redis, or invokes
// the fallback Fetch function on a cache miss, with optional hit/miss hooks.
func FetchOrCompute[T any](ctx context.Context, c *Cache, opts FetchOptions[T]) (Entry[T], error) {
    key := opts.Key

    // 1. Attempt cache lookup if Redis client is available
    if c != nil && c.client != nil {
        var cached T
        hit, err := c.Get(ctx, key, &cached)
        if err == nil && hit {
            if opts.IsValid == nil || opts.IsValid(cached) {
                if opts.OnHit != nil {
                    opts.OnHit(ctx, key, cached)
                }
                return Entry[T]{Data: cached, Key: key, CacheHit: true}, nil
            }
        }
    }

    // 2. Cache miss: execute fallback fetch function
    data, err := opts.Fetch(ctx)
    if err != nil {
        var zero T
        return Entry[T]{Data: zero, Key: key, CacheHit: false}, err
    }

    // 3. Populate Redis cache on valid data
    if c != nil && c.client != nil && (opts.IsValid == nil || opts.IsValid(data)) {
        _ = c.Set(ctx, key, data, opts.TTL)
    }

    // 4. Trigger miss hook
    if opts.OnMiss != nil {
        opts.OnMiss(ctx, key, data)
    }

    return Entry[T]{Data: data, Key: key, CacheHit: false}, nil
}

// Fetch returns the cached value at key if present; otherwise it calls fn,
// caches the result for ttl, and returns it.
func Fetch[T any](ctx context.Context, c *Cache, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
    entry, err := FetchOrCompute(ctx, c, FetchOptions[T]{
        Key: key,
        TTL: ttl,
        Fetch: func(_ context.Context) (T, error) {
            return fn()
        },
    })
    return entry.Data, err
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
