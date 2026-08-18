package cards

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

func parseIDs(c *fiber.Ctx) (eventID, personID uuid.UUID, err error) {
	eventID, err = uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	personID, err = uuid.Parse(c.Params("personId"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	return eventID, personID, nil
}

// Download streams a freshly generated PDF — like the people CSV export
// and the org logo, a deliberate exception to the {"data": ...} envelope
// for a binary response.
func (h *Handler) Download(c *fiber.Ctx) error {
	eventID, personID, err := parseIDs(c)
	if err != nil {
		return err
	}
	pdfBytes, err := h.service.GenerateForPerson(c.Context(), middleware.Claims(c).OrganizationID, eventID, personID)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `inline; filename="id-card.pdf"`)
	return c.Send(pdfBytes)
}

func (h *Handler) Resend(c *fiber.Ctx) error {
	eventID, personID, err := parseIDs(c)
	if err != nil {
		return err
	}
	if err := h.service.ResendForPerson(c.Context(), middleware.Claims(c).OrganizationID, eventID, personID); err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, ResendResponse{Message: "card resent"})
}
