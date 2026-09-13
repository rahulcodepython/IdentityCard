package analytics

import (
    "uuid"
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/utils"
)

func (a *App) handleSummary(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, err := uuid.Parse(c.Params("eventId"))
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgInvalidEventID, err)
    }
    resp, err := a.Summary(c.Context(), orgID, eventID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleOverview(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    resp, err := a.Overview(c.Context(), orgID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDaily(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, err := uuid.Parse(c.Params("eventId"))
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgInvalidEventID, err)
    }

    var subEventID *uuid.UUID
    if raw := c.Query("sub_event_id"); raw != "" {
        id, err := uuid.Parse(raw)
        if err != nil {
            return utils.ErrBadRequest(generic.ErrMsgInvalidSubEventID, err)
        }
        subEventID = &id
    }

    resp, err := a.Daily(c.Context(), orgID, eventID, subEventID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}
