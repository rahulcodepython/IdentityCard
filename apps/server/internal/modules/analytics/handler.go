package analytics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/httpx"
	"identitycard-server/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	resp, err := h.service.Summary(c.Context(), middleware.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Daily(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}

	var subEventID *uuid.UUID
	if raw := c.Query("sub_event_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub_event_id")
		}
		subEventID = &id
	}

	resp, err := h.service.Daily(c.Context(), middleware.Claims(c).OrganizationID, eventID, subEventID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}
