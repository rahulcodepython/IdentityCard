package forms

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts the admin side: /events/:eventId/forms.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)

	g := router.Group("/events/:eventId/forms", middleware.RequireAuth(cfg), middleware.RequireOrganization, manage)
	g.Post("/", h.Create)
	g.Get("/", h.List)
	g.Patch("/:id", h.Update)
	g.Delete("/:id", h.Delete)
}

// RegisterPublicRoutes mounts the unauthenticated side a shared link
// actually points at: /public/forms/:token. Deliberately outside the
// RequireAuth-gated tree entirely — possession of the token is the access
// control.
func RegisterPublicRoutes(router fiber.Router, h *Handler) {
	g := router.Group("/public/forms/:token")
	g.Get("/", h.GetPublic)
	g.Post("/submit", h.Submit)
}
