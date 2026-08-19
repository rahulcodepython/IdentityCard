package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerPlansRoutes mounts the public plan catalog — no auth, since a
// prospective customer needs to see plans before they have an account —
// plus the authenticated subscription endpoints. Subscribing is additive
// (see PlansService.Subscribe): there is no "upgrade" endpoint any more,
// since buying any plan — including a second Custom top-up — is the same
// action, POST /plans/subscriptions.
func registerPlansRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.PlansController) {
	router.Get("/plans", ctrl.List)

	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)
	g := router.Group("/plans", middlewares.RequireAuth(cfg), middlewares.RequireOrganization)
	g.Get("/subscriptions", ctrl.ListSubscriptions)
	g.Post("/subscriptions", manage, ctrl.Subscribe)
	g.Post("/subscriptions/:id/renew", manage, ctrl.RenewSubscription)
}
