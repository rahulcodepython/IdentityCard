package people

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts /events/:eventId/people. Unlike events, every
// route here — including reads — requires admin/super_admin: attendee
// records carry PII (email, mobile) that a scanner-only account has no
// business browsing.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)

	g := router.Group("/events/:eventId/people", middleware.RequireAuth(cfg), manage)
	g.Post("/", h.Create)
	g.Get("/", h.List)
	g.Get("/export", h.Export)
	g.Post("/import", h.Import)
	g.Get("/:id", h.Get)
	g.Patch("/:id", h.Update)
	g.Delete("/:id", h.Delete)
}
