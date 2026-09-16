package eventform

import (
    "github.com/gofiber/fiber/v2"
)

func (h *App) RegisterRoutes(r fiber.Router) {
    group := r.Group("/events/:eventId/form")

    group.Get("/", h.GetEventFormHandler)
    group.Post("/", h.CreateEventFormHandler)
    group.Patch("/", h.UpdateEventFormHandler)
    group.Post("/lock", h.LockEventFormHandler)
    group.Delete("/", h.DeleteEventFormHandler)
}
