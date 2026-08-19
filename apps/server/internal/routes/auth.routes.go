package routes

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	"identitycard-server/internal/middlewares"
)

func registerAuthRoutes(router fiber.Router, cfg *config.Config, ctrl *controllers.AuthController) {
	g := router.Group("/auth")
	g.Post("/register", ctrl.Register)
	g.Post("/otp/send", ctrl.SendOTP)
	g.Post("/otp/verify", ctrl.VerifyOTP)
	g.Post("/totp/verify", ctrl.VerifyTOTP)
	g.Post("/organization", middlewares.RequireAuth(cfg), ctrl.CreateOrganization)
	g.Post("/refresh", ctrl.Refresh)
	g.Post("/logout", ctrl.Logout)
	g.Get("/me", middlewares.RequireAuth(cfg), ctrl.Me)
	// Public — these are full-page browser redirects to/from Google, not
	// XHR calls, so they can't sit behind RequireAuth.
	g.Get("/google/login", ctrl.GoogleLogin)
	g.Get("/google/callback", ctrl.GoogleCallback)
}
