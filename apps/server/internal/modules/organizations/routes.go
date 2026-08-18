package organizations

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes wires the organizations module. GetSettings and
// UploadLogo share the same admin/super_admin gate — the org settings
// page needs both to work for the same roles. GetLogo (rendering the
// image) is open to any authenticated org member.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)

	g := router.Group("/organizations", middleware.RequireAuth(cfg))
	g.Get("/settings", manage, h.GetSettings)
	g.Get("/logo", h.GetLogo)
	g.Post("/logo", manage, h.UploadLogo)
}
