package events

import (
    "context"
    "log"
    "mime/multipart"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
    "identitycard-server/internal/utils"
)

func (a *App) handleCreate(c *fiber.Ctx) error {
    var req CreateEventRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Create(c.Context(), middlewares.Claims(c).OrganizationID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusCreated, resp)
}

func (a *App) handleList(c *fiber.Ctx) error {
    resp, err := a.List(c.Context(), middlewares.Claims(c).OrganizationID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleGet(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    resp, err := a.Get(c.Context(), middlewares.Claims(c).OrganizationID, id)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpdate(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    var req UpdateEventRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Update(c.Context(), middlewares.Claims(c).OrganizationID, id, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handlePublish(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    orgID := middlewares.Claims(c).OrganizationID
    resp, err := a.Publish(c.Context(), orgID, id)
    if err != nil {
        return err
    }

    if a.cardSender != nil {
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("events: panic sending cards for event %s: %v", id, r)
                }
            }()
            a.cardSender.SendForEvent(context.Background(), orgID, id)
        }()
    }

    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDelete(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    if err := a.DeleteDraftEvent(c.Context(), middlewares.Claims(c).OrganizationID, id); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleImportDays(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    file, err := openUploadedCSV(c)
    if err != nil {
        return err
    }
    defer file.Close()

    summary, err := a.ImportDaysCSV(c.Context(), middlewares.Claims(c).OrganizationID, id, file)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, summary)
}

func (a *App) handleExportDays(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    csvBytes, err := a.ExportDays(c.Context(), middlewares.Claims(c).OrganizationID, id)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, generic.ContentTypeCSV)
    c.Set(fiber.HeaderContentDisposition, `attachment; filename="event-days.csv"`)
    return c.Send(csvBytes)
}

func (a *App) handleUploadImage(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    fileHeader, err := c.FormFile("image")
    if err != nil {
        return utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgMissingImageFile)
    }
    file, err := fileHeader.Open()
    if err != nil {
        return utils.ErrInternal()
    }
    defer file.Close()

    contentType := fileHeader.Header.Get("Content-Type")
    if err := a.UploadImage(c.Context(), middlewares.Claims(c).OrganizationID, id, contentType, fileHeader.Size, file); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleGetImage(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    data, contentType, err := a.GetImage(c.Context(), middlewares.Claims(c).OrganizationID, id)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, contentType)
    return c.Send(data)
}

func (a *App) handleUploadOrganizerSignature(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    fileHeader, err := c.FormFile("signature")
    if err != nil {
        return utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgMissingSignatureFile)
    }
    file, err := fileHeader.Open()
    if err != nil {
        return utils.ErrInternal()
    }
    defer file.Close()

    contentType := fileHeader.Header.Get("Content-Type")
    if err := a.UploadOrganizerSignature(c.Context(), middlewares.Claims(c).OrganizationID, id, contentType, fileHeader.Size, file); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleGetOrganizerSignature(c *fiber.Ctx) error {
    id, err := parseEventID(c)
    if err != nil {
        return err
    }
    data, contentType, err := a.GetOrganizerSignature(c.Context(), middlewares.Claims(c).OrganizationID, id)
    if err != nil {
        return err
    }
    c.Set(fiber.HeaderContentType, contentType)
    return c.Send(data)
}

func parseEventID(c *fiber.Ctx) (uuid.UUID, error) {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgInvalidEventID)
    }
    return id, nil
}

func openUploadedCSV(c *fiber.Ctx) (multipart.File, error) {
    fileHeader, err := c.FormFile("file")
    if err != nil {
        return nil, utils.NewError(fiber.StatusBadRequest, generic.ErrCodeBadRequest, generic.ErrMsgMissingCSVFile)
    }
    file, err := fileHeader.Open()
    if err != nil {
        return nil, utils.ErrInternal()
    }
    return file, nil
}
