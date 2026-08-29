package forms

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := router.Group("/events/:eventId/forms", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
    g.Post("/", a.handleCreate)
    g.Get("/", a.handleList)
    g.Patch("/:id", a.handleUpdate)
    g.Delete("/:id", a.handleDelete)
}

func (a *App) RegisterPublicRoutes(router fiber.Router) {
    g := router.Group("/public/forms/:token")
    g.Get("/", a.handleGetPublic)
    g.Post("/submit", a.handleSubmit)
}
