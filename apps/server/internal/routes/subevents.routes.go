package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerSubEventsRoutes mounts /events/:eventId/subevents, nested under
// the events group so every operation is naturally scoped to its parent
// event.
func registerSubEventsRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.SubEventsController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/events/:eventId/subevents", middlewares.RequireAuth(cfg), middlewares.RequireOrganization)
	g.Post("/", manage, ctrl.Create)
	g.Get("/", ctrl.List)
	g.Get("/:id", ctrl.Get)
	g.Patch("/:id", manage, ctrl.Update)
	g.Delete("/:id", manage, ctrl.Delete)
}
