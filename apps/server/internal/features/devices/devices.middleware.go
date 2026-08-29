package devices

import "github.com/gofiber/fiber/v2"

const deviceClaimsLocalsKey = "device_claims"
const KeyHeader = "X-Device-Key"

func (a *App) RequireDevice() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := a.Authenticate(c.Context(), c.Get(KeyHeader))
		if err != nil {
			return err
		}
		c.Locals(deviceClaimsLocalsKey, claims)
		return c.Next()
	}
}

func GetDeviceClaims(c *fiber.Ctx) DeviceClaims {
	claims, _ := c.Locals(deviceClaimsLocalsKey).(DeviceClaims)
	return claims
}
