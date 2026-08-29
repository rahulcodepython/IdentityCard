package cards

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
    "identitycard-server/internal/utils"
)

func (a *App) handleDownload(c *fiber.Ctx) error {
    eventID, personID, err := parseCardsIDs(c)
    if err != nil {
        return err
    }
    pdfBytes, err := a.GenerateForPerson(c.Context(), middlewares.Claims(c).OrganizationID, eventID, personID)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, generic.ContentTypePDF)
    c.Set(fiber.HeaderContentDisposition, `inline; filename="id-card.pdf"`)
    return c.Send(pdfBytes)
}

func (a *App) handleResend(c *fiber.Ctx) error {
    eventID, personID, err := parseCardsIDs(c)
    if err != nil {
        return err
    }
    if err := a.ResendForPerson(c.Context(), middlewares.Claims(c).OrganizationID, eventID, personID); err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, fiber.Map{"message": generic.MsgCardResent})
}

func parseCardsIDs(c *fiber.Ctx) (eventID, personID uuid.UUID, err error) {
    eventID, err = uuid.Parse(c.Params("eventId"))
    if err != nil {
        return uuid.UUID{}, uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgInvalidEventID)
    }
    personID, err = uuid.Parse(c.Params("personId"))
    if err != nil {
        return uuid.UUID{}, uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgInvalidPersonID)
    }
    return eventID, personID, nil
}
