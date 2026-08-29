package routes

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/config"
    "identitycard-server/internal/controllers"
    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

// registerPlansRoutes mounts the public plan catalog (no auth) plus the
// authenticated billing/purchase endpoints. Every write requires admin or
// super_admin role.
func registerPlansRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.PlansController) {
    router.Get("/plans", ctrl.List)

    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)
    g := router.Group("/plans", middlewares.RequireAuth, middlewares.RequireOrganization)
    g.Get("/billing", ctrl.ListBilling)
    g.Post("/purchase", manage, ctrl.Purchase)
    g.Post("/billing/:lineageRootId/renew", manage, ctrl.Renew)
    g.Post("/billing/:lineageRootId/upgrade", manage, ctrl.Upgrade)
    g.Post("/billing/:lineageRootId/cancel", manage, ctrl.Cancel)
}
