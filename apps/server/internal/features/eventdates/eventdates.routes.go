package eventdates

import (
    "github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers HTTP routes for event dates under /events/:eventId/dates.
func (h *App) RegisterRoutes(r fiber.Router) {
    group := r.Group("/events/:eventId/dates")

    group.Get("/", h.ListHandler)
    group.Post("/bulk", h.BulkSaveHandler)
    group.Post("/overwride", h.OverwrideHandler)
    group.Delete("/bulk", h.BulkDeleteHandler)
}
