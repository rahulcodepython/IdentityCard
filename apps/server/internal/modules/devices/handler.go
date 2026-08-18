package devices

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

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateDeviceRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Create(c.Context(), middleware.Claims(c).OrganizationID, req.Name)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusCreated, resp)
}

func (h *Handler) List(c *fiber.Ctx) error {
	resp, err := h.service.List(c.Context(), middleware.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Revoke(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid device id")
	}
	if err := h.service.Revoke(c.Context(), middleware.Claims(c).OrganizationID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Pair is public — no session exists yet for a device that hasn't paired.
func (h *Handler) Pair(c *fiber.Ctx) error {
	var req PairDeviceRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Pair(c.Context(), req.OTPCode)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

// Me is device-key authenticated (see RequireDevice), not session auth —
// it's how the scanner UI shows "connected to <org>" after pairing.
func (h *Handler) Me(c *fiber.Ctx) error {
	claims := GetClaims(c)
	resp, err := h.service.Me(c.Context(), claims.OrganizationID, claims.DeviceID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}
