package plans

import "github.com/gofiber/fiber/v2"

// RegisterRoutes mounts the public plan catalog — no auth, since a
// prospective customer needs to see plans before they have an account.
func RegisterRoutes(router fiber.Router, h *Handler) {
	router.Get("/plans", h.List)
}
