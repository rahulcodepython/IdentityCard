package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerEventsRoutes mounts /events. Reading is open to any authenticated
// org member (e.g. a future scanner UI listing events); every write
// requires admin or super_admin.
func registerEventsRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.EventsController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/events", middlewares.RequireAuth(cfg), middlewares.RequireOrganization)
	g.Post("/", manage, ctrl.Create)
	g.Get("/", ctrl.List)
	g.Get("/:id", ctrl.Get)
	g.Patch("/:id", manage, ctrl.Update)
	g.Post("/:id/publish", manage, ctrl.Publish)
	g.Delete("/:id", manage, ctrl.Delete)

	g.Post("/:id/days/exclusions", manage, ctrl.AddExcludedDate)
	g.Post("/:id/days/import", manage, ctrl.ImportDays)
	g.Post("/:id/days/exclusions/import", manage, ctrl.ImportExcludedDates)
	g.Get("/:id/days/export", ctrl.ExportDays)
}
