package forms

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

func parseEventID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	return id, nil
}

func (h *Handler) Create(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	var req CreateFormRequest
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
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	resp, err := h.service.List(c.Context(), middleware.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid form id")
	}
	var req UpdateFormRequest
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
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid form id")
	}
	if err := h.service.Delete(c.Context(), middleware.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetPublic and Submit serve the unauthenticated /public/forms/:token
// routes — token possession is the only access control.

func (h *Handler) GetPublic(c *fiber.Ctx) error {
	resp, err := h.service.GetPublic(c.Context(), c.Params("token"))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Submit(c *fiber.Ctx) error {
	var req SubmitFormRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.service.Submit(c.Context(), c.Params("token"), req); err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusCreated, SubmitResponse{Message: "registered"})
}
