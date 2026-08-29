package devices

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
)

func (a *App) RequireDevice() fiber.Handler {
    return func(c *fiber.Ctx) error {
        claims, err := a.Authenticate(c.Context(), c.Get(generic.HeaderDeviceKey))
        if err != nil {
            return err
        }
        c.Locals(generic.ContextKeyDeviceClaims, claims)
        return c.Next()
    }
}

func GetDeviceClaims(c *fiber.Ctx) DeviceClaims {
    claims, _ := c.Locals(generic.ContextKeyDeviceClaims).(DeviceClaims)
    return claims
}
