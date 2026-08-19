package services

import "github.com/gofiber/fiber/v2"

const deviceClaimsLocalsKey = "device_claims"

const KeyHeader = "X-Device-Key"

// RequireDevice is a Fiber middleware factory, not a stateless function
// like middlewares.RequireAuth — validating a device key means a DB lookup
// (hash comparison), so it has to live on DevicesService rather than in
// the generic internal/middlewares package.
func (s *DevicesService) RequireDevice() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := s.Authenticate(c.Context(), c.Get(KeyHeader))
		if err != nil {
			return err
		}
		c.Locals(deviceClaimsLocalsKey, claims)
		return c.Next()
	}
}

// GetDeviceClaims returns the authenticated device's claims. Must only be
// called after RequireDevice has run for the route.
func GetDeviceClaims(c *fiber.Ctx) DeviceClaims {
	claims, _ := c.Locals(deviceClaimsLocalsKey).(DeviceClaims)
	return claims
}
