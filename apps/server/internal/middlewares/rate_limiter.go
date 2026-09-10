package middlewares

import (
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/utils"
)

// RateLimiterMiddleware creates a rate limiter bounded per client IP.
// Health check endpoints are exempted so container probes never trigger 429s.
func RateLimiterMiddleware() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        100,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return utils.ErrTooManyRequests(generic.ErrMsgTooManyRequests, nil)
        },
        Next: func(c *fiber.Ctx) bool {
            path := c.Path()
            return path == "/health" || strings.HasSuffix(path, "/health")
        },
    })
}
