package devices

import "github.com/gofiber/fiber/v2"

const claimsLocalsKey = "device_claims"

const KeyHeader = "X-Device-Key"

// RequireDevice is a Fiber middleware factory, not a stateless function
// like middleware.RequireAuth — validating a device key means a DB lookup
// (hash comparison), so it has to live on Service rather than in the
// generic internal/middleware package.
func (s *Service) RequireDevice() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := s.Authenticate(c.Context(), c.Get(KeyHeader))
		if err != nil {
			return err
		}
		c.Locals(claimsLocalsKey, claims)
		return c.Next()
	}
}

// GetClaims returns the authenticated device's claims. Must only be
// called after RequireDevice has run for the route.
func GetClaims(c *fiber.Ctx) Claims {
	claims, _ := c.Locals(claimsLocalsKey).(Claims)
	return claims
}
