package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerOrganizationsRoutes wires GetSettings/UploadLogo behind the same
// admin/super_admin gate — the org settings page needs both to work for
// the same roles. GetLogo (rendering the image) is open to any
// authenticated org member.
func registerOrganizationsRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.OrganizationsController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/organizations", middlewares.RequireAuth, middlewares.RequireOrganization)
	g.Get("/settings", manage, ctrl.GetSettings)
	g.Patch("/settings", manage, ctrl.UpdateSettings)
	g.Delete("/", manage, ctrl.DeleteOrganization)
	g.Get("/logo", ctrl.GetLogo)
	g.Post("/logo", manage, ctrl.UploadLogo)
	g.Delete("/logo", manage, ctrl.DeleteLogo)
}
