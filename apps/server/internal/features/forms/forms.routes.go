package forms

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := protected.Group("/events/:eventId/forms", middlewares.RequireOrganization, manage)

    g.Post("/", a.handleCreate)
    g.Get("/", a.handleList)
    g.Patch("/:id", a.handleUpdate)
    g.Delete("/:id", a.handleDelete)

    pub := public.Group("/public/forms/:token")
    pub.Get("/", a.handleGetPublic)
    pub.Post("/submit", a.handleSubmit)
}
