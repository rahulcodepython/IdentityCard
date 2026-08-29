package plans

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    router.Get("/plans", a.handleListPlans)

    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)
    g := router.Group("/plans", middlewares.RequireAuth, middlewares.RequireOrganization)
    g.Get("/billing", a.handleListBilling)
    g.Get("/subscriptions", a.handleListBilling)
    g.Post("/purchase", manage, a.handlePurchase)
    g.Post("/subscriptions", manage, a.handlePurchase)
    g.Post("/billing/:lineageRootId/renew", manage, a.handleRenew)
    g.Post("/billing/:lineageRootId/upgrade", manage, a.handleUpgrade)
    g.Post("/billing/:lineageRootId/cancel", manage, a.handleCancel)
}
