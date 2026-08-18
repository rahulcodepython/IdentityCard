package analytics

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts /events/:eventId/analytics/{summary,daily}.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)
	g := router.Group("/events/:eventId/analytics", middleware.RequireAuth(cfg), manage)
	g.Get("/summary", h.Summary)
	g.Get("/daily", h.Daily)
}
