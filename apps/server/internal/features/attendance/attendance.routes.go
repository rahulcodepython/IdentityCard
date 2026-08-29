package attendance

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)
    g := router.Group("/events/:eventId/attendance", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
    g.Get("/", a.handleListForEvent)
    g.Get("/export", a.handleExport)
}

func (a *App) RegisterScanRoute(scannerGroup fiber.Router) {
    scannerGroup.Post("/scan", a.handleScan)
}
