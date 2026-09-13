package forms

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleOwner, generic.RoleMember)

    g := protected.Group("/events/:eventId/forms", manage)

    g.Post("/", a.handleCreate)
    g.Get("/", a.handleList)
    g.Patch("/:id", a.handleUpdate)
    g.Delete("/:id", a.handleDelete)

    // Public registration forms: support canonical /public/forms/:token and alias /forms/public/:token
    for _, prefix := range []string{"/public/forms/:token", "/forms/public/:token"} {
        pub := public.Group(prefix)
        pub.Get("/", a.handleGetPublic)
        pub.Post("/submit", a.handleSubmit)
    }
}
