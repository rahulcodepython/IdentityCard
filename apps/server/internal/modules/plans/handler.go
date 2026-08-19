package plans

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

func (h *Handler) List(c *fiber.Ctx) error {
	resp, err := h.service.List(c.Context())
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) ListSubscriptions(c *fiber.Ctx) error {
	resp, err := h.service.ListForOrganization(c.Context(), middleware.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Subscribe(c *fiber.Ctx) error {
	var req SubscribeRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Subscribe(c.Context(), middleware.Claims(c).OrganizationID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusCreated, resp)
}

func (h *Handler) RenewSubscription(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid subscription id")
	}
	resp, err := h.service.Renew(c.Context(), middleware.Claims(c).OrganizationID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}
