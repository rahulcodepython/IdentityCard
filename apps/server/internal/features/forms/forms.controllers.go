package forms

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
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    var req CreateFormRequest
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
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    resp, err := a.List(c.Context(), orgID, eventID)
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
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgInvalidFormID, err)
    }
    var req UpdateFormRequest
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
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgInvalidFormID, err)
    }
    if err := a.Delete(c.Context(), orgID, eventID, id); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleGetPublic(c *fiber.Ctx) error {
    resp, err := a.GetPublic(c.Context(), c.Params("token"))
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleSubmit(c *fiber.Ctx) error {
    var req SubmitFormRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    if err := a.Submit(c.Context(), c.Params("token"), req); err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusCreated, fiber.Map{"message": generic.MsgRegistered})
}

func parseFormsEventID(c *fiber.Ctx) (uuid.UUID, error) {
    id, err := uuid.Parse(c.Params("eventId"))
    if err != nil {
        return uuid.UUID{}, utils.ErrBadRequest(generic.ErrMsgInvalidEventID, err)
    }
    return id, nil
}
