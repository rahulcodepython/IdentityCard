package utils

import (
    "strconv"

    "github.com/gofiber/fiber/v2"
)

// CursorParams reads the "after_id"/"before_id"/"limit" query params shared
// by every cursor-paginated feed (notifications, logs, security events).
// afterID means "what's new since this id" (initial load, poll, refresh);
// beforeID means "give me older ones before this id" (Load More). At most
// one of the two is ever set by a well-behaved client — if both are present,
// beforeID wins.
func CursorParams(c *fiber.Ctx) (afterID, beforeID *int64, limit int) {
    limit, _ = strconv.Atoi(c.Query("limit", "10"))
    if limit < 1 {
        limit = 1
    }
    if limit > 100 {
        limit = 100
    }

    if v := c.Query("before_id"); v != "" {
        if id, err := strconv.ParseInt(v, 10, 64); err == nil {
            beforeID = &id
            return
        }
    }
    if v := c.Query("after_id"); v != "" {
        if id, err := strconv.ParseInt(v, 10, 64); err == nil {
            afterID = &id
        }
    }
    return
}
