package cards

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts /events/:eventId/people/:personId/card. Same PII
// sensitivity call as people itself: admin/super_admin only, no plain
// authenticated read.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)

	g := router.Group("/events/:eventId/people/:personId/card", middleware.RequireAuth(cfg), middleware.RequireOrganization, manage)
	g.Get("/", h.Download)
	g.Post("/resend", h.Resend)
}
