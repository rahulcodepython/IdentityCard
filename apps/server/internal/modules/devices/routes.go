package devices

import (
	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

// RegisterRoutes mounts the admin side: /devices. super_admin only — per
// the spec, managing who/what can act on the org (roles, devices) is a
// super_admin-exclusive concern, distinct from admin/super_admin managing
// events themselves.
func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	g := router.Group("/devices", middleware.RequireAuth(cfg), middleware.RequireOrganization, middleware.RequireRole(coreauth.RoleSuperAdmin))
	g.Post("/", h.Create)
	g.Get("/", h.List)
	g.Post("/:id/revoke", h.Revoke)
}

// RegisterPublicRoutes mounts /public/devices/pair — a device has no
// session until this succeeds, so it can't be behind RequireAuth.
func RegisterPublicRoutes(router fiber.Router, h *Handler) {
	router.Post("/public/devices/pair", h.Pair)
}

// RegisterScannerRoutes mounts the device-key-authenticated /scanner
// group's identity endpoint. The scan endpoint itself is mounted by
// attendance.RegisterRoutes on the same group — see cmd/api/main.go,
// which is the only place that needs both.
func RegisterScannerRoutes(router fiber.Router, s *Service, h *Handler) fiber.Router {
	g := router.Group("/scanner", s.RequireDevice())
	g.Get("/me", h.Me)
	return g
}
