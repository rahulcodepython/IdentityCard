package register

import (
    "github.com/gofiber/fiber/v2"
)

func (h *App) RegisterRoutes(r fiber.Router) {
    group := r.Group("/public/apply")

    group.Get("/:eventFormId", h.GetPublicApplyConfigHandler)
    group.Post("/:eventFormId", h.SubmitApplicationHandler)
}
