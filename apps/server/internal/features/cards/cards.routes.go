package cards

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := router.Group("/events/:eventId/people/:personId/card", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
    g.Get("/", a.handleDownload)
    g.Post("/resend", a.handleResend)
}
