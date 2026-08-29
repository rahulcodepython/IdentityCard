package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
)

// registerDevicesRoutes mounts the admin side: /devices. super_admin
// only — per the spec, managing who/what can act on the org (roles,
// devices) is a super_admin-exclusive concern, distinct from
// admin/super_admin managing events themselves.
func registerDevicesRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.DevicesController) {
	g := router.Group("/devices", middlewares.RequireAuth, middlewares.RequireOrganization, middlewares.RequireRole(generic.RoleSuperAdmin))
	g.Post("/", ctrl.Create)
	g.Get("/", ctrl.List)
	g.Post("/:id/revoke", ctrl.Revoke)
}

// registerDevicesPublicRoutes mounts /public/devices/pair — a device has
// no session until this succeeds, so it can't be behind RequireAuth.
func registerDevicesPublicRoutes(router fiber.Router, ctrl *controllers.DevicesController) {
	router.Post("/public/devices/pair", ctrl.Pair)
}

// registerDevicesScannerRoutes mounts the device-key-authenticated
// /scanner group's identity endpoint. The scan endpoint itself is mounted
// by registerAttendanceScanRoute on the same returned group — see
// routes.go's SetUp, the one place that needs both.
func registerDevicesScannerRoutes(router fiber.Router, svc *services.DevicesService, ctrl *controllers.DevicesController) fiber.Router {
	g := router.Group("/scanner", svc.RequireDevice())
	g.Get("/me", ctrl.Me)
	return g
}
