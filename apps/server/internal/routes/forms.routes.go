package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerFormsRoutes mounts the admin side: /events/:eventId/forms.
func registerFormsRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.FormsController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/events/:eventId/forms", middlewares.RequireAuth(cfg), middlewares.RequireOrganization, manage)
	g.Post("/", ctrl.Create)
	g.Get("/", ctrl.List)
	g.Patch("/:id", ctrl.Update)
	g.Delete("/:id", ctrl.Delete)
}

// registerFormsPublicRoutes mounts the unauthenticated side a shared link
// actually points at: /public/forms/:token. Deliberately outside the
// RequireAuth-gated tree entirely — possession of the token is the access
// control.
func registerFormsPublicRoutes(router fiber.Router, ctrl *controllers.FormsController) {
	g := router.Group("/public/forms/:token")
	g.Get("/", ctrl.GetPublic)
	g.Post("/submit", ctrl.Submit)
}
