package devices

import "github.com/gofiber/fiber/v2"

func (h *App) RegisterRoutes(api fiber.Router) {
    h.RegisterGlobalDeviceRoutes(api)
    h.RegisterEventDeviceRoutes(api)
}
