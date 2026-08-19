package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type AttendanceController struct {
	service *services.AttendanceService
}

func NewAttendanceController(service *services.AttendanceService) *AttendanceController {
	return &AttendanceController{service: service}
}

// Scan is device-key authenticated (see services.DevicesService.RequireDevice),
// not session auth — a scanner bot never logs in as an org member.
func (ctrl *AttendanceController) Scan(c *fiber.Ctx) error {
	var req entities.ScanRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	claims := services.GetDeviceClaims(c)
	resp, err := ctrl.service.Scan(c.Context(), claims.OrganizationID, claims.DeviceID, req.QRToken)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *AttendanceController) ListForEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	filter, err := attendanceFilterFromQuery(c)
	if err != nil {
		return err
	}
	resp, err := ctrl.service.BuildRoster(c.Context(), middlewares.Claims(c).OrganizationID, eventID, filter)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

// Export returns a CSV file directly — the same deliberate exception to
// the {"data": ...} envelope used by PeopleController.Export and friends.
func (ctrl *AttendanceController) Export(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	filter, err := attendanceFilterFromQuery(c)
	if err != nil {
		return err
	}
	csvBytes, err := ctrl.service.Export(c.Context(), middlewares.Claims(c).OrganizationID, eventID, filter)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="attendance.csv"`)
	return c.Send(csvBytes)
}

func attendanceFilterFromQuery(c *fiber.Ctx) (services.RosterFilter, error) {
	var filter services.RosterFilter
	if raw := c.Query("sub_event_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return services.RosterFilter{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub_event_id")
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
