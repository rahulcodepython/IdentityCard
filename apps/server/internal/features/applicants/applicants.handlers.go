package applicants

import (
    "encoding/json"
    "errors"
    "log/slog"

    "github.com/gofiber/fiber/v2"
    "identitycard-server/internal/utils"
)

func (h *App) GetApplicantSchemaHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")

    schema, err := h.GetApplicantSchemaService(c.UserContext(), eventID)
    if err != nil {
        slog.Error("failed to get applicant schema", "error", err, "eventId", eventID)
        if errors.Is(err, ErrInvalidEventID) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to get applicant schema", err)
    }

    return utils.OK(c, "applicant schema retrieved successfully", schema)
}

func (h *App) ListApplicantsHandler(c *fiber.Ctx) error {
    eventID := c.Params("eventId")
    page, limit := utils.PaginationParams(c)
    if limit <= 0 || limit > 100 {
        limit = 30
    }
    search := c.Query("search")

    var filters []ApplicantFilter
    filtersParam := c.Query("filters")
    if filtersParam != "" {
        _ = json.Unmarshal([]byte(filtersParam), &filters)
    }

    if field := c.Query("field"); field != "" {
        filters = append(filters, ApplicantFilter{
            Field: field,
            Op:    c.Query("op", "eq"),
            Value: c.Query("value", c.Query("val")),
        })
    }

    res, err := h.ListApplicantsService(c.UserContext(), eventID, search, filters, page, limit)
    if err != nil {
        slog.Error("failed to list applicants", "error", err, "eventId", eventID, "filters", filters)
        if errors.Is(err, ErrInvalidEventID) {
            return utils.ErrBadRequest(c, err.Error(), err)
        }
        return utils.ErrInternal(c, "failed to list applicants", err)
    }

    return utils.OK(c, "applicants retrieved successfully", res)
}
