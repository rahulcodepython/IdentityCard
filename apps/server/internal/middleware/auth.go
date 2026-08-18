package middleware

import (
	"slices"

	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/httpx"
)

const claimsLocalsKey = "auth_claims"

// RequireAuth verifies the access-token cookie and, on success, stores the
// parsed claims on the request context for downstream handlers/middleware
// to read via Claims(c).
func RequireAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(auth.AccessCookieName)
		if token == "" {
			return httpx.ErrUnauthorized("")
		}
		claims, err := auth.ParseAccessToken(cfg.JWTSecret, token)
		if err != nil {
			return httpx.ErrUnauthorized("")
		}
		c.Locals(claimsLocalsKey, claims)
		return c.Next()
	}
}

// Claims returns the authenticated request's claims. It must only be
// called after RequireAuth has run for the route.
func Claims(c *fiber.Ctx) *auth.AccessClaims {
	claims, _ := c.Locals(claimsLocalsKey).(*auth.AccessClaims)
	return claims
}

// RequireRole grants access if the caller holds ANY of the given roles.
// Must run after RequireAuth.
func RequireRole(roles ...auth.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := Claims(c)
		if claims == nil {
			return httpx.ErrUnauthorized("")
		}
		if slices.ContainsFunc(roles, claims.HasRole) {
			return c.Next()
		}
		return httpx.ErrForbidden("")
	}
}
