package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerCardsRoutes mounts /events/:eventId/people/:personId/card. Same
// PII sensitivity call as people itself: admin/super_admin only, no plain
// authenticated read.
func registerCardsRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.CardsController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/events/:eventId/people/:personId/card", middlewares.RequireAuth(cfg), middlewares.RequireOrganization, manage)
	g.Get("/", ctrl.Download)
	g.Post("/resend", ctrl.Resend)
}
