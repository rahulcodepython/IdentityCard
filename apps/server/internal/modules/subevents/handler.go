package subevents

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

func parseIDParams(c *fiber.Ctx) (eventID, subEventID uuid.UUID, err error) {
	eventID, err = uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	if idParam := c.Params("id"); idParam != "" {
		subEventID, err = uuid.Parse(idParam)
		if err != nil {
			return uuid.UUID{}, uuid.UUID{}, httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub-event id")
		}
	}
	return eventID, subEventID, nil
}

func (h *Handler) Create(c *fiber.Ctx) error {
	eventID, _, err := parseIDParams(c)
	if err != nil {
		return err
	}
	var req CreateSubEventRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Create(c.Context(), middleware.Claims(c).OrganizationID, eventID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusCreated, resp)
}

func (h *Handler) List(c *fiber.Ctx) error {
	eventID, _, err := parseIDParams(c)
	if err != nil {
		return err
	}
	resp, err := h.service.List(c.Context(), middleware.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	eventID, id, err := parseIDParams(c)
	if err != nil {
		return err
	}
	resp, err := h.service.Get(c.Context(), middleware.Claims(c).OrganizationID, eventID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	eventID, id, err := parseIDParams(c)
	if err != nil {
		return err
	}
	var req UpdateSubEventRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Update(c.Context(), middleware.Claims(c).OrganizationID, eventID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	eventID, id, err := parseIDParams(c)
	if err != nil {
		return err
	}
	if err := h.service.Delete(c.Context(), middleware.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
