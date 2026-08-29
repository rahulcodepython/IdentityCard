package forms

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
    "identitycard-server/internal/utils"
)

func (a *App) handleCreate(c *fiber.Ctx) error {
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    var req CreateFormRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Create(c.Context(), middlewares.Claims(c).OrganizationID, eventID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusCreated, resp)
}

func (a *App) handleList(c *fiber.Ctx) error {
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    resp, err := a.List(c.Context(), middlewares.Claims(c).OrganizationID, eventID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpdate(c *fiber.Ctx) error {
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgInvalidFormID)
    }
    var req UpdateFormRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Update(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDelete(c *fiber.Ctx) error {
    eventID, err := parseFormsEventID(c)
    if err != nil {
        return err
    }
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgInvalidFormID)
    }
    if err := a.Delete(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id); err != nil {
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
        return uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgInvalidEventID)
    }
    return id, nil
}
