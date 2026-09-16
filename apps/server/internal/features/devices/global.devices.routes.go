package devices

import "github.com/gofiber/fiber/v2"

func (h *App) RegisterGlobalDeviceRoutes(api fiber.Router) {
    devicesGroup := api.Group("/devices")
    devicesGroup.Post("/", h.CreateDeviceHandler)
    devicesGroup.Get("/", h.ListDevicesHandler)
    devicesGroup.Get("/me", h.GetMyDeviceHandler)
    devicesGroup.Post("/verify", h.VerifyDeviceHandler)
    devicesGroup.Post("/webauthn/register-options", h.WebAuthnRegisterOptionsHandler)
    devicesGroup.Post("/webauthn/register-verify", h.WebAuthnRegisterVerifyHandler)
    devicesGroup.Post("/webauthn/login-options", h.WebAuthnLoginOptionsHandler)
    devicesGroup.Post("/webauthn/login-verify", h.WebAuthnLoginVerifyHandler)
    devicesGroup.Patch("/:id", h.UpdateDeviceHandler)
    devicesGroup.Post("/:id/regenerate-pin", h.RegeneratePINHandler)
    devicesGroup.Delete("/:id", h.DeleteDeviceHandler)
}
