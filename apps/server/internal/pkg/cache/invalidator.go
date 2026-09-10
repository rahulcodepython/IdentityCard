package cache

import (
    "context"
    "fmt"
    "log/slog"
)

// Invalidate purges the cache keys matching each given pattern (e.g. "courses:*").
func (c *Cache) Invalidate(ctx context.Context, patterns ...string) {
    for _, p := range patterns {
        slog.Info("invalidating cache", "pattern", p)
        _ = c.DeleteByPattern(ctx, p)
    }
}

// InvalidateWishlist purges wishlist cache entries for a user (or every
// wishlist entry when userID is empty) — kept as a dedicated method since
// the pattern itself depends on the argument, unlike every other domain.
func (c *Cache) InvalidateWishlist(ctx context.Context, userID string) {
    if userID != "" {
        c.Invalidate(ctx, fmt.Sprintf("wishlist:user:%s:*", userID))
    } else {
        c.Invalidate(ctx, "wishlist:*")
    }
}

// AuthCacheKey builds the per-user roles/permissions cache key. Kept here so
// middlewares.BaseAuthMiddleware (which reads it) and every invalidation call
// site (which writes it) agree on the exact key format without either one
// importing the other.
func AuthCacheKey(userID string) string {
    return fmt.Sprintf("auth:roles_permissions:%s", userID)
}

// InvalidateUserAuthCache purges the cached roles/permissions for one user —
// call after a targeted mutation (assigning/revoking roles for that user).
func (c *Cache) InvalidateUserAuthCache(ctx context.Context, userID string) {
    slog.Info("invalidating auth cache", "user_id", userID)
    _ = c.Delete(ctx, AuthCacheKey(userID))
}

// InvalidateAllUserAuthCache purges every cached per-user roles/permissions
// entry — call after a role-definition change (permissions attached to a
// role), since it's unknown which users currently hold that role.
func (c *Cache) InvalidateAllUserAuthCache(ctx context.Context) {
    slog.Info("invalidating all per-user auth cache")
    _ = c.DeleteByPattern(ctx, "auth:roles_permissions:*")
}
