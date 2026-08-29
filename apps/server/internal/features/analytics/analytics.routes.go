package analytics

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := router.Group("/events/:eventId/analytics", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
    g.Get("/summary", a.handleSummary)
    g.Get("/daily", a.handleDaily)

    router.Get("/analytics/overview", middlewares.RequireAuth, middlewares.RequireOrganization, manage, a.handleOverview)
}
