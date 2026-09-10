package cache

import (
    "context"
    "encoding/json"
    "time"
)

// GraceTokenPayload holds rotated token information during grace period.
type GraceTokenPayload struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

// AcquireLock attempts to acquire a Redis lock for 5 seconds to prevent race conditions during token rotation.
func (c *Cache) AcquireLock(ctx context.Context, lockKey string) (bool, error) {
    if c == nil || c.client == nil {
        return true, nil
    }
    return c.client.SetNX(ctx, "lock:"+lockKey, "1", 5*time.Second).Result()
}

// ReleaseLock releases the acquired Redis lock.
func (c *Cache) ReleaseLock(ctx context.Context, lockKey string) {
    if c != nil && c.client != nil {
        _ = c.client.Del(ctx, "lock:"+lockKey).Err()
    }
}

// GetGraceTokens retrieves recently rotated tokens during parallel race conditions.
func (c *Cache) GetGraceTokens(ctx context.Context, oldRefreshToken string) (*GraceTokenPayload, bool) {
    if c == nil || c.client == nil {
        return nil, false
    }
    val, err := c.client.Get(ctx, "grace:"+oldRefreshToken).Result()
    if err != nil {
        return nil, false
    }
    var payload GraceTokenPayload
    if err := json.Unmarshal([]byte(val), &payload); err != nil {
        return nil, false
    }
    return &payload, true
}

// SetGraceTokens caches newly rotated tokens for 30 seconds to serve parallel requests seamlessly.
func (c *Cache) SetGraceTokens(ctx context.Context, oldRefreshToken string, payload *GraceTokenPayload) {
    if c == nil || c.client == nil {
        return
    }
    data, err := json.Marshal(payload)
    if err != nil {
        return
    }
    _ = c.client.Set(ctx, "grace:"+oldRefreshToken, data, 30*time.Second).Err()
}
