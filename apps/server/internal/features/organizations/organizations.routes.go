package organizations

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)
    adminOnly := middlewares.RequireRole(generic.RoleAdmin)

    g := router.Group("/organizations", middlewares.RequireAuth, middlewares.RequireOrganization)
    g.Get("/settings", manage, a.handleGetSettings)
    g.Patch("/settings", manage, a.handleUpdateSettings)
    g.Delete("/", adminOnly, a.handleDeleteOrganization)
    g.Get("/logo", a.handleGetLogo)
    g.Post("/logo", manage, a.handleUploadLogo)
    g.Delete("/logo", manage, a.handleDeleteLogo)
}
