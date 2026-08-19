package subevents

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts /events/:eventId/subevents, nested under the
// events group so every operation is naturally scoped to its parent event.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)

	g := router.Group("/events/:eventId/subevents", middleware.RequireAuth(cfg), middleware.RequireOrganization)
	g.Post("/", manage, h.Create)
	g.Get("/", h.List)
	g.Get("/:id", h.Get)
	g.Patch("/:id", manage, h.Update)
	g.Delete("/:id", manage, h.Delete)
}
