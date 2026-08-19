package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type CardsController struct {
	service *services.CardsService
}

func NewCardsController(service *services.CardsService) *CardsController {
	return &CardsController{service: service}
}

func parseCardsIDs(c *fiber.Ctx) (eventID, personID uuid.UUID, err error) {
	eventID, err = uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	personID, err = uuid.Parse(c.Params("personId"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	return eventID, personID, nil
}

// Download streams a freshly generated PDF — like the people CSV export
// and the org logo, a deliberate exception to the {"data": ...} envelope
// for a binary response.
func (ctrl *CardsController) Download(c *fiber.Ctx) error {
	eventID, personID, err := parseCardsIDs(c)
	if err != nil {
		return err
	}
	pdfBytes, err := ctrl.service.GenerateForPerson(c.Context(), middlewares.Claims(c).OrganizationID, eventID, personID)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `inline; filename="id-card.pdf"`)
	return c.Send(pdfBytes)
}

func (ctrl *CardsController) Resend(c *fiber.Ctx) error {
	eventID, personID, err := parseCardsIDs(c)
	if err != nil {
		return err
	}
	if err := ctrl.service.ResendForPerson(c.Context(), middlewares.Claims(c).OrganizationID, eventID, personID); err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, entities.MessageResponse{Message: "card resent"})
}
