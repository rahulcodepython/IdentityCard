package plans

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts the public plan catalog — no auth, since a
// prospective customer needs to see plans before they have an account —
// plus the authenticated subscription endpoints. Subscribing is additive
// (see Service.Subscribe): there is no "upgrade" endpoint any more, since
// buying any plan — including a second Custom top-up — is the same
// action, POST /plans/subscriptions.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	router.Get("/plans", h.List)

	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)
	g := router.Group("/plans", middleware.RequireAuth(cfg), middleware.RequireOrganization)
	g.Get("/subscriptions", h.ListSubscriptions)
	g.Post("/subscriptions", manage, h.Subscribe)
	g.Post("/subscriptions/:id/renew", manage, h.RenewSubscription)
}
