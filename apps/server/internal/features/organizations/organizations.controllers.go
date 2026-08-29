package organizations

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
    "identitycard-server/internal/utils"
)

func (a *App) handleGetSettings(c *fiber.Ctx) error {
    claims := middlewares.Claims(c)
    resp, err := a.GetSettings(c.Context(), claims.OrganizationID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpdateSettings(c *fiber.Ctx) error {
    var req UpdateOrganizationSettingsRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    claims := middlewares.Claims(c)
    resp, err := a.UpdateSettings(c.Context(), claims.OrganizationID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDeleteOrganization(c *fiber.Ctx) error {
    claims := middlewares.Claims(c)
    if err := a.DeleteOrganization(c.Context(), claims.OrganizationID); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleDeleteLogo(c *fiber.Ctx) error {
    claims := middlewares.Claims(c)
    if err := a.DeleteLogo(c.Context(), claims.OrganizationID); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleUploadLogo(c *fiber.Ctx) error {
    fileHeader, err := c.FormFile("logo")
    if err != nil {
        return utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgMissingLogoFile)
    }
    file, err := fileHeader.Open()
    if err != nil {
        return utils.ErrInternal()
    }
    defer file.Close()

    contentType := fileHeader.Header.Get("Content-Type")
    claims := middlewares.Claims(c)
    if err := a.UploadLogo(c.Context(), claims.OrganizationID, contentType, fileHeader.Size, file); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleGetLogo(c *fiber.Ctx) error {
    claims := middlewares.Claims(c)
    data, contentType, err := a.GetLogo(c.Context(), claims.OrganizationID)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, contentType)
    return c.Send(data)
}
