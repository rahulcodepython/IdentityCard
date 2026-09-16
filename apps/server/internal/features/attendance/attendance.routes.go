package attendance

import (
    "github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all attendance and analysis routes.
func (h *App) RegisterRoutes(r fiber.Router) {
    attendanceGroup := r.Group("/events/:eventId/attendance")
    attendanceGroup.Post("/scan", h.ScanApplicantHandler)
    attendanceGroup.Post("/entry", h.MarkEntryHandler)
    attendanceGroup.Post("/exit", h.MarkExitHandler)

    analysisGroup := r.Group("/events/:eventId/analysis")
    analysisGroup.Get("/metrics", h.GetAttendanceMetricsHandler)
    analysisGroup.Get("/attendees", h.ListAttendeeAnalysisHandler)
}
