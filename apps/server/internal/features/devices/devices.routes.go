package devices

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := router.Group("/devices", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
    g.Post("/", a.handleCreate)
    g.Get("/", a.handleList)
    g.Post("/:id/revoke", a.handleRevoke)
}

func (a *App) RegisterPublicRoutes(router fiber.Router) {
    router.Post("/public/devices/pair", a.handlePair)
}

func (a *App) RegisterScannerRoutes(router fiber.Router) fiber.Router {
    g := router.Group("/scanner", a.RequireDevice())
    g.Get("/me", a.handleMe)
    return g
}
