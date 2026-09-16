package devices

import "github.com/gofiber/fiber/v2"

func (h *App) RegisterEventDeviceRoutes(api fiber.Router) {
    eventDevicesGroup := api.Group("/events/:eventId/devices")
    eventDevicesGroup.Get("/", h.ListEventDevicesHandler)
    eventDevicesGroup.Get("/available", h.ListAvailableGlobalDevicesHandler)
    eventDevicesGroup.Post("/assign", h.AssignEventDevicesHandler)
    eventDevicesGroup.Delete("/:deviceId", h.UnassignEventDeviceHandler)
}
