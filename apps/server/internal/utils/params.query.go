package utils

import (
    "github.com/gofiber/fiber/v2"
)

// RequireQuery reads a required query-string parameter, returning a 400
// APIError if it's missing/empty instead of leaving every controller to
// hand-write the same check.
func RequireQuery(c *fiber.Ctx, key, label string) (string, error) {
    val := c.Query(key)
    if val == "" {
        return "", ErrBadRequest(label+" query param required.", nil)
    }
    return val, nil
}
