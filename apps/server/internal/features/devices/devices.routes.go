package devices

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := protected.Group("/devices", middlewares.RequireOrganization, manage)

    g.Post("/", a.handleCreate)
    g.Get("/", a.handleList)
    g.Post("/:id/revoke", a.handleRevoke)

    public.Post("/public/devices/pair", a.handlePair)

    scanner := public.Group("/scanner", a.RequireDevice())
    scanner.Get("/me", a.handleMe)
}
