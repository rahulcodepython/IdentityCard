package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type DevicesController struct {
	service *services.DevicesService
}

func NewDevicesController(service *services.DevicesService) *DevicesController {
	return &DevicesController{service: service}
}

func (ctrl *DevicesController) Create(c *fiber.Ctx) error {
	var req entities.CreateDeviceRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Create(c.Context(), middlewares.Claims(c).OrganizationID, req.Name)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (ctrl *DevicesController) List(c *fiber.Ctx) error {
	resp, err := ctrl.service.List(c.Context(), middlewares.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *DevicesController) Revoke(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid device id")
	}
	if err := ctrl.service.Revoke(c.Context(), middlewares.Claims(c).OrganizationID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Pair is public — no session exists yet for a device that hasn't paired.
func (ctrl *DevicesController) Pair(c *fiber.Ctx) error {
	var req entities.PairDeviceRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Pair(c.Context(), req.OTPCode)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

// Me is device-key authenticated (see services.DevicesService.RequireDevice),
// not session auth — it's how the scanner UI shows "connected to <org>"
// after pairing.
func (ctrl *DevicesController) Me(c *fiber.Ctx) error {
	claims := services.GetDeviceClaims(c)
	resp, err := ctrl.service.Me(c.Context(), claims.OrganizationID, claims.DeviceID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}
