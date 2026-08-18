package attendance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/httpx"
	"identitycard-server/internal/middleware"
	"identitycard-server/internal/modules/devices"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Scan is device-key authenticated (see devices.RequireDevice), not
// session auth — a scanner bot never logs in as an org member.
func (h *Handler) Scan(c *fiber.Ctx) error {
	var req ScanRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	claims := devices.GetClaims(c)
	resp, err := h.service.Scan(c.Context(), claims.OrganizationID, claims.DeviceID, req.QRToken)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) ListForEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	filter, err := filterFromQuery(c)
	if err != nil {
		return err
	}
	resp, err := h.service.BuildRoster(c.Context(), middleware.Claims(c).OrganizationID, eventID, filter)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

// Export returns a CSV file directly — the same deliberate exception to
// the {"data": ...} envelope used by people.Handler.Export and friends.
func (h *Handler) Export(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	filter, err := filterFromQuery(c)
	if err != nil {
		return err
	}
	csvBytes, err := h.service.Export(c.Context(), middleware.Claims(c).OrganizationID, eventID, filter)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="attendance.csv"`)
	return c.Send(csvBytes)
}

func filterFromQuery(c *fiber.Ctx) (RosterFilter, error) {
	var filter RosterFilter
	if raw := c.Query("sub_event_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return RosterFilter{}, httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub_event_id")
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
