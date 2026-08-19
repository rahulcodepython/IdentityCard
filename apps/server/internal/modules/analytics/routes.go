package analytics

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts /events/:eventId/analytics/{summary,daily} and the
// org-wide /analytics/overview used by the dashboard home page.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)

	g := router.Group("/events/:eventId/analytics", middleware.RequireAuth(cfg), middleware.RequireOrganization, manage)
	g.Get("/summary", h.Summary)
	g.Get("/daily", h.Daily)

	router.Get("/analytics/overview", middleware.RequireAuth(cfg), middleware.RequireOrganization, manage, h.Overview)
}
