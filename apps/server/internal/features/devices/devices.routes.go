package devices

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleOwner, generic.RoleMember)

    g := protected.Group("/devices", manage)

    g.Post("/", a.handleCreate)
    g.Get("/", a.handleList)
    g.Delete("/:id", a.handleRevoke)
    g.Post("/:id/revoke", a.handleRevoke)

    // Device pairing: canonical /public/devices/pair and alias /devices/pair
    public.Post("/public/devices/pair", a.handlePair)
    public.Post("/devices/pair", a.handlePair)

    scanner := public.Group("/scanner", a.RequireDevice())
    scanner.Get("/me", a.handleMe)
}
