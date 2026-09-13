package organizations

import (
    "uuid"

    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
    "identitycard-server/internal/utils"
)

func (a *App) handleListOrganizations(c *fiber.Ctx) error {
    claims := middlewares.Claims(c)
    if claims == nil {
        return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
    }
    userUUID, err := claims.UserID()
    if err != nil || userUUID == uuid.Nil() {
        return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
    }

    resp, err := a.ListOrganizations(c.Context(), userUUID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleGetSettings(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    resp, err := a.GetSettings(c.Context(), orgID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpdateSettings(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    var req UpdateOrganizationSettingsRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.UpdateSettings(c.Context(), orgID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDeleteOrganization(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    claims := middlewares.Claims(c)
    if claims == nil {
        return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
    }
    userUUID, err := claims.UserID()
    if err != nil || userUUID == uuid.Nil() {
        return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
    }
    if err := a.DeleteOrganization(c.Context(), orgID, userUUID); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}
