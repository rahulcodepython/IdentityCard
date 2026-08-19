package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

func registerMembersRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.MembersController) {
	members := router.Group("/organizations/members", middlewares.RequireAuth(cfg), middlewares.RequireOrganization)

	members.Get("/", middlewares.RequireRole(generic.RoleAdmin), ctrl.List)
	members.Delete("/:id", middlewares.RequireRole(generic.RoleAdmin), ctrl.Delete)
}
