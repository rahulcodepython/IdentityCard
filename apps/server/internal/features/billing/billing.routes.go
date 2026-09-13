package billing

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    ownerOnly := middlewares.RequireRole(generic.RoleOwner)

    g := protected.Group("/billing", ownerOnly)

    g.Get("/", a.handleGetBilling)
    g.Post("/credits/purchase", a.handlePurchaseCredits)
    g.Post("/renew", a.handleRenewAnnual)
}
