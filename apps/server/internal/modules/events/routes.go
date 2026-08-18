package events

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts /events. Reading is open to any authenticated org
// member (e.g. a future scanner UI listing events); every write requires
// admin or super_admin.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)

	g := router.Group("/events", middleware.RequireAuth(cfg))
	g.Post("/", manage, h.Create)
	g.Get("/", h.List)
	g.Get("/:id", h.Get)
	g.Patch("/:id", manage, h.Update)
	g.Post("/:id/publish", manage, h.Publish)
	g.Delete("/:id", manage, h.Delete)
}
