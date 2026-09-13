package organizations

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, userProtected fiber.Router) {
    ownerOnly := middlewares.RequireRole(generic.RoleOwner)

    // User-scoped organization listing
    userProtected.Get("/organizations", a.handleListOrganizations)

    // Organization-scoped settings and deletion
    protected.Get("/settings", ownerOnly, a.handleGetSettings)
    protected.Patch("/settings", ownerOnly, a.handleUpdateSettings)
    protected.Delete("/", ownerOnly, a.handleDeleteOrganization)
}
