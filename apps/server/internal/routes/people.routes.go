package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerPeopleRoutes mounts /events/:eventId/people. Unlike events,
// every route here — including reads — requires admin/super_admin:
// attendee records carry PII (email, mobile) that a scanner-only account
// has no business browsing.
func registerPeopleRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.PeopleController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/events/:eventId/people", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
	g.Post("/", ctrl.Create)
	g.Get("/", ctrl.List)
	g.Get("/export", ctrl.Export)
	g.Post("/import", ctrl.Import)
	g.Get("/:id", ctrl.Get)
	g.Patch("/:id", ctrl.Update)
	g.Delete("/:id", ctrl.Delete)
}
