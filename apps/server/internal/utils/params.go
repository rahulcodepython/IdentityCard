package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams extracts sanitized page and limit integers from query parameters.
func PaginationParams(c *fiber.Ctx) (int, int) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}

// RequireQuery reads a required query string parameter, returning an error if missing or empty.
func RequireQuery(c *fiber.Ctx, key, label string) (string, error) {
	val := c.Query(key)
	if val == "" {
		return "", ErrBadRequest(c, label+" query param required", nil)
	}
	return val, nil
}
