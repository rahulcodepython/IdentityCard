package events

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

func (a *App) RegisterRoutes(protected, public fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := protected.Group("/events", middlewares.RequireOrganization)

    g.Post("/", manage, a.handleCreate)
    g.Get("/", a.handleList)
    g.Get("/:id", a.handleGet)
    g.Patch("/:id", manage, a.handleUpdate)
    g.Post("/:id/publish", manage, a.handlePublish)
    g.Delete("/:id", manage, a.handleDelete)

    // Day schedule endpoints (standard events only for import).
    g.Post("/:id/days/import", manage, a.handleImportDays)
    g.Get("/:id/days/export", a.handleExportDays)

    // Event image upload/download.
    g.Patch("/:id/image", manage, a.handleUploadImage)
    g.Get("/:id/image", a.handleGetImage)

    // Organizer signature upload/download.
    g.Patch("/:id/organizer-signature", manage, a.handleUploadOrganizerSignature)
    g.Get("/:id/organizer-signature", a.handleGetOrganizerSignature)
}
