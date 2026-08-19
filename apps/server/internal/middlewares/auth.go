package middlewares

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/config"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/pkg/jwt"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

const claimsLocalsKey = "auth_claims"

var authService *services.AuthService

// InitAuth initializes the auth middleware with the auth service for token refreshing.
func InitAuth(service *services.AuthService) {
	authService = service
}

// RequireAuth verifies the `ic_access` cookie and, on success, stores the parsed 
// claims on the request context for downstream handlers/middleware to read.
// If the access token is expired but `ic_refresh` is valid, it automatically 
// issues new cookies and allows the request to continue.
func RequireAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Cookies("ic_access")
		refreshToken := c.Cookies("ic_refresh")

		if accessToken == "" && refreshToken == "" {
			return utils.ErrUnauthorized("missing authentication")
		}

		// Try parsing access token first
		claims, err := jwt.ParseAccessToken(cfg.JWTSecret, accessToken)
		if err == nil {
			c.Locals(claimsLocalsKey, claims)
			return c.Next()
		}

		// If access token is invalid or expired, check if we have a refresh token
		if refreshToken == "" || authService == nil {
			return utils.ErrUnauthorized("expired authentication")
		}

		// Attempt auto-refresh
		newAccess, newRefresh, accessTTL, refreshTTL, err := authService.Refresh(c.Context(), refreshToken)
		if err != nil {
			// Refresh failed, clear cookies and reject
			c.Cookie(&fiber.Cookie{Name: "ic_access", Value: "", MaxAge: -1, Path: "/"})
			c.Cookie(&fiber.Cookie{Name: "ic_refresh", Value: "", MaxAge: -1, Path: "/"})
			return utils.ErrUnauthorized("session expired")
		}

		// Set new cookies
		c.Cookie(&fiber.Cookie{
			Name:     "ic_access",
			Value:    newAccess,
			HTTPOnly: true,
			Secure:   cfg.Env == "production",
			SameSite: "Lax",
			MaxAge:   int(accessTTL.Seconds()),
			Path:     "/",
		})
		c.Cookie(&fiber.Cookie{
			Name:     "ic_refresh",
			Value:    newRefresh,
			HTTPOnly: true,
			Secure:   cfg.Env == "production",
			SameSite: "Lax",
			MaxAge:   int(refreshTTL.Seconds()),
			Path:     "/",
		})

		// Parse the newly issued access token to set claims
		newClaims, err := jwt.ParseAccessToken(cfg.JWTSecret, newAccess)
		if err != nil {
			return utils.ErrUnauthorized("failed to parse refreshed token")
		}
		
		c.Locals(claimsLocalsKey, newClaims)
		return c.Next()
	}
}

// Claims returns the authenticated request's claims. It must only be
// called after RequireAuth has run for the route.
func Claims(c *fiber.Ctx) *jwt.AccessClaims {
	claims, _ := c.Locals(claimsLocalsKey).(*jwt.AccessClaims)
	return claims
}

// RequireRole grants access if the caller holds ANY of the given roles.
// Must run after RequireAuth.
func RequireRole(roles ...generic.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := Claims(c)
		if claims == nil {
			return utils.ErrUnauthorized("")
		}
		if slices.ContainsFunc(roles, claims.HasRole) {
			return c.Next()
		}
		return utils.ErrForbidden("")
	}
}

// RequireOrganization additionally requires the caller's access token to
// carry an organization. A signed-in user who hasn't finished onboarding
// yet (a Google signup — see services.AuthService.CreateOrganization)
// has none, and every business route needs one to scope its queries by.
// Must run after RequireAuth.
func RequireOrganization(c *fiber.Ctx) error {
	claims := Claims(c)
	if claims == nil || claims.OrganizationID == uuid.Nil {
		return utils.NewError(fiber.StatusForbidden, "organization_required", "finish onboarding first")
	}
	return c.Next()
}
