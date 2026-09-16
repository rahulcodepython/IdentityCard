package forms

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers HTTP routes for the forms resource under /forms.
func (h *App) RegisterRoutes(r fiber.Router) {
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
