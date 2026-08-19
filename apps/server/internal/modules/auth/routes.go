package auth

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/middleware"
)

func RegisterRoutes(router fiber.Router, cfg *config.Config, h *Handler) {
	g := router.Group("/auth")
	g.Post("/register", h.Register)
	g.Post("/otp/send", h.SendOTP)
	g.Post("/otp/verify", h.VerifyOTP)
	g.Post("/totp/verify", h.VerifyTOTP)
	g.Post("/organization", middleware.RequireAuth(cfg), h.CreateOrganization)
	g.Post("/refresh", h.Refresh)
	g.Post("/logout", h.Logout)
	g.Get("/me", middleware.RequireAuth(cfg), h.Me)
	// Public — these are full-page browser redirects to/from Google, not
	// XHR calls, so they can't sit behind the cookie-based RequireAuth.
	g.Get("/google/login", h.GoogleLogin)
	g.Get("/google/callback", h.GoogleCallback)
}
