package middlewares

import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"
)

// RequestContextMiddleware establishes a per-request context derived from the
// server's lifecycle context and bounded by a timeout deadline. It assigns this context
// to Fiber via c.SetUserContext(ctx).
//
// Any downstream database queries (pgx), cache commands (redis), or outbound
// operations using c.UserContext() will automatically receive cancellation if:
//  1. The request duration exceeds the timeout deadline (context.DeadlineExceeded).
//  2. The server begins graceful shutdown (context.Canceled).
func RequestContextMiddleware(rootCtx context.Context, timeout time.Duration) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx, cancel := context.WithTimeout(rootCtx, timeout)
        defer cancel()

        c.SetUserContext(ctx)
        return c.Next()
    }
}
