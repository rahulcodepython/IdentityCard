package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
)

// registerAttendanceScanRoute mounts POST /scan onto the already
// device-authenticated group registerDevicesScannerRoutes returns — only
// routes.go's SetUp needs both, so it's the one place that wires them
// together.
func registerAttendanceScanRoute(scannerGroup fiber.Router, ctrl *controllers.AttendanceController) {
	scannerGroup.Post("/scan", ctrl.Scan)
}

// registerAttendanceRoutes mounts the admin-facing reads:
// /events/:eventId/attendance (the roster, filterable via query params)
// and its CSV export.
func registerAttendanceRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.AttendanceController) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)
	g := router.Group("/events/:eventId/attendance", middlewares.RequireAuth(cfg), middlewares.RequireOrganization, manage)
	g.Get("/", ctrl.ListForEvent)
	g.Get("/export", ctrl.Export)
}
