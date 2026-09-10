package utils

import (
    "strconv"

    "github.com/gofiber/fiber/v2"
)

// maxPage bounds the page number so `(page-1) * limit` (computed downstream
// by every repository's offset calculation) can never overflow int on a
// 64-bit build even at the max limit (100) — comfortably beyond any real
// pagination depth.
const maxPage = 1_000_000_000

// PaginationParams extracts sanitized page and limit integers from query parameters.
func PaginationParams(c *fiber.Ctx) (int, int) {
    page, err := strconv.Atoi(c.Query("page", "1"))
    if err != nil || page < 1 {
        page = 1
    }
    if page > maxPage {
        page = maxPage
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
