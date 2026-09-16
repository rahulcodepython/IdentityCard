package eventsharing

import (
    "errors"

    "github.com/gofiber/fiber/v2"
    "identitycard-server/internal/utils"
)

func (h *App) GetSharingHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")

    res, err := h.GetSharingService(c.UserContext(), eventID)
    if err != nil {
        if errors.Is(err, ErrInvalidEventID) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to get event sharing status", err)
    }

    return utils.OK(c, "event sharing retrieved successfully", res)
}

func (h *App) UpdateSharingHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")
    req, err := utils.ParseBody[UpdateEventSharingRequest](c.Body())
    if err != nil {
        return utils.ErrBadRequest(c, err.Error(), err)
    }

    res, err := h.UpdateSharingService(c.UserContext(), eventID, *req)
    if err != nil {
        if errors.Is(err, ErrInvalidEventID) || errors.Is(err, ErrInvalidFormID) ||
            errors.Is(err, ErrFormNotPublished) || errors.Is(err, ErrCannotChangeForm) ||
            errors.Is(err, ErrInvalidMaxApplicants) || errors.Is(err, ErrInvalidExpiryDate) ||
            errors.Is(err, ErrInvalidStatus) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to update event sharing configuration", err)
    }

    return utils.OK(c, "event sharing configured successfully", res)
}

