package dates

import (
    "errors"

    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/utils"
)

// ListHandler returns all event dates for an event, optionally filtered by month (YYYY-MM).
func (h *App) ListHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")
    month := c.Query("month")

    res, err := h.ListByMonthService(c.UserContext(), eventID, month)
    if err != nil {
        if errors.Is(err, ErrInvalidDateFormat) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to retrieve event dates", err)
    }

    return utils.OK(c, "event dates retrieved successfully", res)
}

// BulkSaveHandler atomically validates and upserts event dates.
func (h *App) BulkSaveHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")

    req, err := utils.ParseBody[BulkUpsertEventDatesRequest](c.Body())
    if err != nil {
        return utils.ErrBadRequest(c, err.Error(), err)
    }

    res, err := h.BulkUpsertService(c.UserContext(), eventID, *req)
    if err != nil {
        if errors.Is(err, ErrEventNotFound) {
            return utils.ErrNotFound(c, err.Error(), err)
        }
        if errors.Is(err, ErrDateOutOfRange) ||
            errors.Is(err, ErrInvalidDateFormat) ||
            errors.Is(err, ErrInvalidTimeFormat) ||
            errors.Is(err, ErrInvalidTimeOrder) ||
            errors.Is(err, ErrEmptyDates) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to save event dates", err)
    }

    return utils.OK(c, "event dates saved successfully", res)
}

// OverwrideHandler deletes all existing event dates and inserts the provided batch.
func (h *App) OverwrideHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")

    req, err := utils.ParseBody[BulkUpsertEventDatesRequest](c.Body())
    if err != nil {
        return utils.ErrBadRequest(c, err.Error(), err)
    }

    res, err := h.OverwrideService(c.UserContext(), eventID, *req)
    if err != nil {
        if errors.Is(err, ErrEventNotFound) {
            return utils.ErrNotFound(c, err.Error(), err)
        }
        if errors.Is(err, ErrDateOutOfRange) ||
            errors.Is(err, ErrInvalidDateFormat) ||
            errors.Is(err, ErrInvalidTimeFormat) ||
            errors.Is(err, ErrInvalidTimeOrder) ||
            errors.Is(err, ErrEmptyDates) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to overwride event dates", err)
    }

    return utils.OK(c, "event dates overwridden successfully", res)
}


// BulkDeleteHandler removes multiple event dates at once.
func (h *App) BulkDeleteHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")

    req, err := utils.ParseBody[BulkDeleteEventDatesRequest](c.Body())
    if err != nil {
        return utils.ErrBadRequest(c, err.Error(), err)
    }

    err = h.BulkDeleteService(c.UserContext(), eventID, *req)
    if err != nil {
        if errors.Is(err, ErrInvalidDateFormat) || errors.Is(err, ErrEmptyDates) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to delete event dates", err)
    }

    return utils.OKEmpty(c, "event dates deleted successfully")
}
