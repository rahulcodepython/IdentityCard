package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type PeopleController struct {
	service *services.PeopleService
}

func NewPeopleController(service *services.PeopleService) *PeopleController {
	return &PeopleController{service: service}
}

func parsePeopleEventID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	return id, nil
}

func (ctrl *PeopleController) Create(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	var req entities.CreatePersonRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Create(c.Context(), middlewares.Claims(c).OrganizationID, eventID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (ctrl *PeopleController) List(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	resp, err := ctrl.service.List(c.Context(), middlewares.Claims(c).OrganizationID, eventID, peopleFilterFromQuery(c))
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *PeopleController) Get(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	resp, err := ctrl.service.Get(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *PeopleController) Update(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	var req entities.UpdatePersonRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Update(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *PeopleController) Delete(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	if err := ctrl.service.Delete(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *PeopleController) Import(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}

	file, err := openUploadedCSV(c)
	if err != nil {
		return err
	}
	defer file.Close()

	var opts services.PeopleImportOptions
	if raw := c.FormValue("sub_event_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub_event_id")
		}
		opts.SubEventID = &id
	}

	summary, err := ctrl.service.ImportCSV(c.Context(), middlewares.Claims(c).OrganizationID, eventID, file, opts)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, summary)
}

// Export returns a CSV file directly — a deliberate exception to the
// standard {"data": ...} envelope, since the response is a file download.
func (ctrl *PeopleController) Export(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}

	var ids []uuid.UUID
	if raw := c.Query("ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			id, err := uuid.Parse(strings.TrimSpace(part))
			if err != nil {
				return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid id in ids")
			}
			ids = append(ids, id)
		}
	}

	csvBytes, err := ctrl.service.Export(c.Context(), middlewares.Claims(c).OrganizationID, eventID, peopleFilterFromQuery(c), ids)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="people.csv"`)
	return c.Send(csvBytes)
}

func peopleFilterFromQuery(c *fiber.Ctx) services.PeopleListFilter {
	var filter services.PeopleListFilter
	if raw := c.Query("sub_event_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			filter.SubEventID = &id
		}
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}
	return filter
}
