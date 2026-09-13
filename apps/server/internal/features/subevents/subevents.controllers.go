package subevents

import (
    "uuid"
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/utils"
)

func (a *App) handleCreate(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, _, err := parseSubEventIDParams(c)
    if err != nil {
        return err
    }
    var req CreateSubEventRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Create(c.Context(), orgID, eventID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusCreated, resp)
}

func (a *App) handleList(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, _, err := parseSubEventIDParams(c)
    if err != nil {
        return err
    }
    resp, err := a.List(c.Context(), orgID, eventID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleGet(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, id, err := parseSubEventIDParams(c)
    if err != nil {
        return err
    }
    resp, err := a.Get(c.Context(), orgID, eventID, id)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpdate(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, id, err := parseSubEventIDParams(c)
    if err != nil {
        return err
    }
    var req UpdateSubEventRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Update(c.Context(), orgID, eventID, id, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDelete(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, id, err := parseSubEventIDParams(c)
    if err != nil {
        return err
    }
    if err := a.Delete(c.Context(), orgID, eventID, id); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func parseSubEventIDParams(c *fiber.Ctx) (eventID, subEventID uuid.UUID, err error) {
    eventID, err = uuid.Parse(c.Params("eventId"))
    if err != nil {
        return uuid.UUID{}, uuid.UUID{}, utils.ErrBadRequest(generic.ErrMsgInvalidEventID, err)
    }
    if idParam := c.Params("id"); idParam != "" {
        subEventID, err = uuid.Parse(idParam)
        if err != nil {
            return uuid.UUID{}, uuid.UUID{}, utils.ErrBadRequest(generic.ErrMsgInvalidSubEventID, err)
        }
    }
    return eventID, subEventID, nil
}
