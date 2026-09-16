package eventsharing

import (
    "github.com/gofiber/fiber/v2"
)

func (h *App) RegisterRoutes(r fiber.Router) {
    group := r.Group("/events/:eventId/sharing")

    group.Get("/", h.GetSharingHandler)
    group.Patch("/", h.UpdateSharingHandler)
}
