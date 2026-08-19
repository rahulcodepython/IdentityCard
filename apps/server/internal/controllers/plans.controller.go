package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type PlansController struct {
	service *services.PlansService
}

func NewPlansController(service *services.PlansService) *PlansController {
	return &PlansController{service: service}
}

func (ctrl *PlansController) List(c *fiber.Ctx) error {
	resp, err := ctrl.service.List(c.Context())
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *PlansController) ListSubscriptions(c *fiber.Ctx) error {
	resp, err := ctrl.service.ListForOrganization(c.Context(), middlewares.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *PlansController) Subscribe(c *fiber.Ctx) error {
	var req entities.SubscribeRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Subscribe(c.Context(), middlewares.Claims(c).OrganizationID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (ctrl *PlansController) RenewSubscription(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid subscription id")
	}
	resp, err := ctrl.service.Renew(c.Context(), middlewares.Claims(c).OrganizationID, id)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}
