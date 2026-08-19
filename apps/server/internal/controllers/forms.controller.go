package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type FormsController struct {
	service *services.FormsService
}

func NewFormsController(service *services.FormsService) *FormsController {
	return &FormsController{service: service}
}

func parseFormsEventID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	return id, nil
}

func (ctrl *FormsController) Create(c *fiber.Ctx) error {
	eventID, err := parseFormsEventID(c)
	if err != nil {
		return err
	}
	var req entities.CreateFormRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Create(c.Context(), middlewares.Claims(c).OrganizationID, eventID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (ctrl *FormsController) List(c *fiber.Ctx) error {
	eventID, err := parseFormsEventID(c)
	if err != nil {
		return err
	}
	resp, err := ctrl.service.List(c.Context(), middlewares.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *FormsController) Update(c *fiber.Ctx) error {
	eventID, err := parseFormsEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid form id")
	}
	var req entities.UpdateFormRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Update(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *FormsController) Delete(c *fiber.Ctx) error {
	eventID, err := parseFormsEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid form id")
	}
	if err := ctrl.service.Delete(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetPublic and Submit serve the unauthenticated /public/forms/:token
// routes — token possession is the only access control.

func (ctrl *FormsController) GetPublic(c *fiber.Ctx) error {
	resp, err := ctrl.service.GetPublic(c.Context(), c.Params("token"))
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *FormsController) Submit(c *fiber.Ctx) error {
	var req entities.SubmitFormRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.service.Submit(c.Context(), c.Params("token"), req); err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, entities.MessageResponse{Message: "registered"})
}
