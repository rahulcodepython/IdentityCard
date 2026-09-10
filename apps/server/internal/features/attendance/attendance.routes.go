package attendance

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)
    g := protected.Group("/events/:eventId/attendance", middlewares.RequireOrganization, manage)

    g.Get("/", a.handleListForEvent)
    g.Get("/export", a.handleExport)

    scanner := public.Group("/scanner", a.devices.RequireDevice())
    scanner.Post("/scan", a.handleScan)
}
