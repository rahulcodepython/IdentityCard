package attendance

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterScanRoute mounts POST /scan onto the already device-authenticated
// group devices.RegisterScannerRoutes returns — only cmd/api/main.go needs
// both packages, so it's the one place that wires them together.
func RegisterScanRoute(scannerGroup fiber.Router, h *Handler) {
	scannerGroup.Post("/scan", h.Scan)
}

// RegisterRoutes mounts the admin-facing reads: /events/:eventId/attendance
// (the roster, filterable via query params — see filterFromQuery) and its
// CSV export.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	manage := middleware.RequireRole(coreauth.RoleAdmin, coreauth.RoleSuperAdmin)
	g := router.Group("/events/:eventId/attendance", middleware.RequireAuth(cfg), manage)
	g.Get("/", h.ListForEvent)
	g.Get("/export", h.Export)
}
