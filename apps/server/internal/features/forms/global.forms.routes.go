package forms

import "github.com/gofiber/fiber/v2"

func (h *App) RegisterGlobalFormRoutes(r fiber.Router) {
    group := r.Group("/forms")

    group.Get("/", h.ListHandler)
    group.Post("/", h.CreateHandler)
    group.Get("/:id", h.GetHandler)
    group.Put("/:id", h.UpdateHandler)
    group.Patch("/:id", h.UpdateHandler)
    group.Put("/:id/fields", h.UpdateFieldsHandler)
    group.Patch("/:id/fields", h.UpdateFieldsHandler)
    group.Delete("/:id", h.DeleteHandler)
}
