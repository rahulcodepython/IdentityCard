package subevents

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := protected.Group("/events/:eventId/subevents", middlewares.RequireOrganization)

    g.Post("/", manage, a.handleCreate)
    g.Get("/", a.handleList)
    g.Get("/:id", a.handleGet)
    g.Patch("/:id", manage, a.handleUpdate)
    g.Delete("/:id", manage, a.handleDelete)
}
