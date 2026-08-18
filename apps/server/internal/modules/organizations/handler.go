package organizations

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/httpx"
	"identitycard-server/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSettings(c *fiber.Ctx) error {
	claims := middleware.Claims(c)
	resp, err := h.service.GetSettings(c.Context(), claims.OrganizationID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) UploadLogo(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("logo")
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "missing logo file")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return httpx.ErrInternal()
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	claims := middleware.Claims(c)
	if err := h.service.UploadLogo(c.Context(), claims.OrganizationID, contentType, fileHeader.Size, file); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetLogo streams the image directly — like the people CSV export, a
// deliberate exception to the {"data": ...} envelope for a binary response.
func (h *Handler) GetLogo(c *fiber.Ctx) error {
	claims := middleware.Claims(c)
	data, contentType, err := h.service.GetLogo(c.Context(), claims.OrganizationID)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, contentType)
	return c.Send(data)
}
