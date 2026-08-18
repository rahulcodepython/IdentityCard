package auth

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	g := router.Group("/auth")
	g.Post("/register", h.Register)
	g.Post("/login", h.Login)
	g.Post("/refresh", h.Refresh)
	g.Post("/logout", h.Logout)
	g.Get("/me", middleware.RequireAuth(cfg), h.Me)
}
