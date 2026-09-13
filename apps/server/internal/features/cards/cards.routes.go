package cards

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleOwner, generic.RoleMember)

    // Support both canonical plural (/cards) and singular (/card) endpoints
    for _, path := range []string{
        "/events/:eventId/people/:personId/cards",
        "/events/:eventId/people/:personId/card",
    } {
        g := protected.Group(path, manage)
        g.Get("/", a.handleDownload)
        g.Post("/resend", a.handleResend)
    }
}
