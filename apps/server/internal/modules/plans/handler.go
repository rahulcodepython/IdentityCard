package plans

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/httpx"
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
