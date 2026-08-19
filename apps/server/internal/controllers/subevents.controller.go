package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type SubEventsController struct {
	service *services.SubEventsService
}

func NewSubEventsController(service *services.SubEventsService) *SubEventsController {
	return &SubEventsController{service: service}
}

func parseSubEventIDParams(c *fiber.Ctx) (eventID, subEventID uuid.UUID, err error) {
	eventID, err = uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	if idParam := c.Params("id"); idParam != "" {
		subEventID, err = uuid.Parse(idParam)
		if err != nil {
			return uuid.UUID{}, uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub-event id")
		}
	}
	return eventID, subEventID, nil
}

func (ctrl *SubEventsController) Create(c *fiber.Ctx) error {
	eventID, _, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	var req entities.CreateSubEventRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Create(c.Context(), middlewares.Claims(c).OrganizationID, eventID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (ctrl *SubEventsController) List(c *fiber.Ctx) error {
	eventID, _, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	resp, err := ctrl.service.List(c.Context(), middlewares.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *SubEventsController) Get(c *fiber.Ctx) error {
	eventID, id, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	resp, err := ctrl.service.Get(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *SubEventsController) Update(c *fiber.Ctx) error {
	eventID, id, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	var req entities.UpdateSubEventRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Update(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *SubEventsController) Delete(c *fiber.Ctx) error {
	eventID, id, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	if err := ctrl.service.Delete(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
