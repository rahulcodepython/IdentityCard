package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerAnalyticsRoutes mounts /events/:eventId/analytics/{summary,daily}
// and the org-wide /analytics/overview used by the dashboard home page.
func registerAnalyticsRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.AnalyticsController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/events/:eventId/analytics", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
	g.Get("/summary", ctrl.Summary)
	g.Get("/daily", ctrl.Daily)

	router.Get("/analytics/overview", middlewares.RequireAuth, middlewares.RequireOrganization, manage, ctrl.Overview)
}
