package dates

import (
    "github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers HTTP routes for event dates under /events/:eventId/dates.
func (h *App) RegisterRoutes(r fiber.Router) {
    group := r.Group("/events/:eventId/dates")

    group.Get("/", h.ListHandler)
    group.Post("/override", h.OverrideHandler)
    group.Post("/sync", h.SyncHandler)
}
