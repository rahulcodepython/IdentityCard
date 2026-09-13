package events

import (
    "mime/multipart"

    "github.com/gofiber/fiber/v2"
    "uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/utils"
)

func (a *App) handleCreate(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    var req CreateEventRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Create(c.Context(), orgID, req)
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
    resp, err := a.List(c.Context(), orgID)
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
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    resp, err := a.Get(c.Context(), orgID, id)
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
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    var req UpdateEventRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Update(c.Context(), orgID, id, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handlePublish(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    resp, err := a.Publish(c.Context(), orgID, id)
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
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    if err := a.DeleteDraftEvent(c.Context(), orgID, id); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleImportDays(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    file, err := openUploadedCSV(c)
    if err != nil {
        return err
    }
    defer file.Close()

    summary, err := a.ImportDaysCSV(c.Context(), orgID, id, file)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, summary)
}

func (a *App) handleExportDays(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    csvBytes, err := a.ExportDays(c.Context(), orgID, id)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, generic.ContentTypeCSV)
    c.Set(fiber.HeaderContentDisposition, `attachment; filename="event-days.csv"`)
    return c.Send(csvBytes)
}

func (a *App) handleUploadImage(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    fileHeader, err := c.FormFile("image")
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgMissingImageFile, err)
    }
    file, err := fileHeader.Open()
    if err != nil {
        return utils.ErrInternal("Failed to open uploaded image.", err)
    }
    defer file.Close()

    contentType := fileHeader.Header.Get("Content-Type")
    if err := a.UploadImage(c.Context(), orgID, id, contentType, fileHeader.Size, file); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleGetImage(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    data, contentType, err := a.GetImage(c.Context(), orgID, id)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, contentType)
    return c.Send(data)
}

func (a *App) handleUploadOrganizerSignature(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    fileHeader, err := c.FormFile("signature")
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgMissingSignatureFile, err)
    }
    file, err := fileHeader.Open()
    if err != nil {
        return utils.ErrInternal("Failed to open uploaded signature.", err)
    }
    defer file.Close()

    contentType := fileHeader.Header.Get("Content-Type")
    if err := a.UploadOrganizerSignature(c.Context(), orgID, id, contentType, fileHeader.Size, file); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleGetOrganizerSignature(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    data, contentType, err := a.GetOrganizerSignature(c.Context(), orgID, id)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, contentType)
    return c.Send(data)
}

func parseEventID(c *fiber.Ctx) (uuid.UUID, error) {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return uuid.UUID{}, utils.ErrBadRequest(generic.ErrMsgInvalidEventID, err)
    }
    return id, nil
}

func openUploadedCSV(c *fiber.Ctx) (multipart.File, error) {
    fileHeader, err := c.FormFile("file")
    if err != nil {
        return nil, utils.ErrBadRequest(generic.ErrMsgMissingCSVFile, err)
    }
    file, err := fileHeader.Open()
    if err != nil {
        return nil, utils.ErrInternal("Failed to open uploaded CSV.", err)
    }
    return file, nil
}
