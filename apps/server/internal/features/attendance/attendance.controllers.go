package attendance

import (
    "uuid"
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/features/devices"
    "identitycard-server/internal/generic"
    "identitycard-server/internal/utils"
)

func (a *App) handleScan(c *fiber.Ctx) error {
    var req ScanRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    claims := devices.GetDeviceClaims(c)
    resp, err := a.Scan(c.Context(), claims.OrganizationID, claims.DeviceID, req.QRToken)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleListForEvent(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, err := uuid.Parse(c.Params("eventId"))
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgInvalidEventID, err)
    }
    filter, err := attendanceFilterFromQuery(c)
    if err != nil {
        return err
    }
    resp, err := a.BuildRoster(c.Context(), orgID, eventID, filter)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleExport(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }
    eventID, err := uuid.Parse(c.Params("eventId"))
    if err != nil {
        return utils.ErrBadRequest(generic.ErrMsgInvalidEventID, err)
    }
    filter, err := attendanceFilterFromQuery(c)
    if err != nil {
        return err
    }
    csvBytes, err := a.Export(c.Context(), orgID, eventID, filter)
    if err != nil {
        return err
    }

    c.Set(fiber.HeaderContentType, generic.ContentTypeCSV)
    c.Set(fiber.HeaderContentDisposition, `attachment; filename="attendance.csv"`)
    return c.Send(csvBytes)
}

func attendanceFilterFromQuery(c *fiber.Ctx) (RosterFilter, error) {
    var filter RosterFilter
    if raw := c.Query("sub_event_id"); raw != "" {
        id, err := uuid.Parse(raw)
        if err != nil {
            return RosterFilter{}, utils.ErrBadRequest(generic.ErrMsgInvalidSubEventID, err)
        }
        filter.SubEventID = &id
    }
    if date := c.Query("date"); date != "" {
        filter.Date = &date
    }
    if status := c.Query("status"); status != "" {
        filter.Status = &status
    }
    if attended := c.Query("attended"); attended != "" {
        v := attended == "true"
        filter.Attended = &v
    }
    return filter, nil
}
