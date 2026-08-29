package routes

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/config"
    "identitycard-server/internal/controllers"
    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
)

// registerEventsRoutes mounts /events. Reading is open to any authenticated
// org member; every write requires admin or super_admin.
func registerEventsRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.EventsController) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

    g := router.Group("/events", middlewares.RequireAuth, middlewares.RequireOrganization)
    g.Post("/", manage, ctrl.Create)
    g.Get("/", ctrl.List)
    g.Get("/:id", ctrl.Get)
    g.Patch("/:id", manage, ctrl.Update)
    g.Post("/:id/publish", manage, ctrl.Publish)
    g.Delete("/:id", manage, ctrl.Delete)

    // Day schedule endpoints (standard events only for import).
    g.Post("/:id/days/import", manage, ctrl.ImportDays)
    g.Get("/:id/days/export", ctrl.ExportDays)

    // Event image upload/download.
    g.Patch("/:id/image", manage, ctrl.UploadImage)
    g.Get("/:id/image", ctrl.GetImage)

    // Organizer signature upload/download.
    g.Patch("/:id/organizer-signature", manage, ctrl.UploadOrganizerSignature)
    g.Get("/:id/organizer-signature", ctrl.GetOrganizerSignature)
}
