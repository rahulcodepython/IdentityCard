package controllers

import (
	"context"
	"log"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

func parseEventID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	return id, nil
}

// EventsCardSender is implemented by *services.CardsService. Defined here,
// on the consumer side, rather than importing that package's concrete type
// directly — cards itself depends on services.EventsService (to read event
// details), so importing it back here would cycle. internal/routes wires
// the concrete *services.CardsService in, since it's the one place both
// services already exist.
type EventsCardSender interface {
	SendForEvent(ctx context.Context, orgID, eventID uuid.UUID)
}

// openUploadedCSV opens a "file" multipart field — shared by every
// domain's CSV-import controller method (events, people, forms), so
// there's exactly one definition of "missing CSV file" error handling.
func openUploadedCSV(c *fiber.Ctx) (multipart.File, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, utils.NewError(fiber.StatusBadRequest, "bad_request", "missing CSV file")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, utils.ErrInternal()
	}
	return file, nil
}

type EventsController struct {
	service    *services.EventsService
	cardSender EventsCardSender
}

func NewEventsController(service *services.EventsService, cardSender EventsCardSender) *EventsController {
	return &EventsController{service: service, cardSender: cardSender}
}

func (ctrl *EventsController) Create(c *fiber.Ctx) error {
	var req entities.CreateEventRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Create(c.Context(), middlewares.Claims(c).OrganizationID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (ctrl *EventsController) List(c *fiber.Ctx) error {
	resp, err := ctrl.service.List(c.Context(), middlewares.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *EventsController) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	resp, err := ctrl.service.Get(c.Context(), middlewares.Claims(c).OrganizationID, id)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *EventsController) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	var req entities.UpdateEventRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Update(c.Context(), middlewares.Claims(c).OrganizationID, id, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *EventsController) Publish(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	orgID := middlewares.Claims(c).OrganizationID
	resp, err := ctrl.service.Publish(c.Context(), orgID, id)
	if err != nil {
		return err
	}

	// Fire-and-forget: emailing every attendee their card shouldn't block
	// the publish response. This goroutine outlives the request (hence
	// context.Background(), not c.Context()) and MUST recover its own
	// panics — Fiber's recover middleware only guards the request
	// goroutine, not ones spawned from inside a handler; an unrecovered
	// panic here would crash the whole process.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("events: panic sending cards for event %s: %v", id, r)
			}
		}()
		ctrl.cardSender.SendForEvent(context.Background(), orgID, id)
	}()

	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *EventsController) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	if err := ctrl.service.Delete(c.Context(), middlewares.Claims(c).OrganizationID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *EventsController) AddExcludedDate(c *fiber.Ctx) error {
	id, err := parseEventID(c)
	if err != nil {
		return err
	}
	var req entities.AddExcludedDateRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.AddExcludedDate(c.Context(), middlewares.Claims(c).OrganizationID, id, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *EventsController) ImportDays(c *fiber.Ctx) error {
	id, err := parseEventID(c)
	if err != nil {
		return err
	}
	file, err := openUploadedCSV(c)
	if err != nil {
		return err
	}
	defer file.Close()

	summary, err := ctrl.service.ImportDaysCSV(c.Context(), middlewares.Claims(c).OrganizationID, id, file)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, summary)
}

func (ctrl *EventsController) ImportExcludedDates(c *fiber.Ctx) error {
	id, err := parseEventID(c)
	if err != nil {
		return err
	}
	file, err := openUploadedCSV(c)
	if err != nil {
		return err
	}
	defer file.Close()

	summary, err := ctrl.service.ImportExcludedDatesCSV(c.Context(), middlewares.Claims(c).OrganizationID, id, file)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, summary)
}

// ExportDays returns a CSV file directly — same deliberate exception to
// the standard {"data": ...} envelope as PeopleController.Export, for the
// same reason (a file download response).
func (ctrl *EventsController) ExportDays(c *fiber.Ctx) error {
	id, err := parseEventID(c)
	if err != nil {
		return err
	}
	csvBytes, err := ctrl.service.ExportDays(c.Context(), middlewares.Claims(c).OrganizationID, id)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="event-days.csv"`)
	return c.Send(csvBytes)
}
