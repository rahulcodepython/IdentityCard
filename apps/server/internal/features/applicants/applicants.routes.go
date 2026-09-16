package applicants

import (
    "github.com/gofiber/fiber/v2"
)

func (h *App) RegisterRoutes(r fiber.Router) {
    group := r.Group("/events/:eventId/applicants")

    group.Get("/schema", h.GetApplicantSchemaHandler)
    group.Get("/", h.ListApplicantsHandler)
}
