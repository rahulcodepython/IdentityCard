package analytics

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := protected.Group("/events/:eventId/analytics", middlewares.RequireOrganization, manage)
    g.Get("/summary", a.handleSummary)
    g.Get("/daily", a.handleDaily)

    protected.Get("/analytics/overview", middlewares.RequireOrganization, manage, a.handleOverview)
}
