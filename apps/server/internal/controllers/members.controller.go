package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type MembersController struct {
	service *services.MembersService
}

func NewMembersController(service *services.MembersService) *MembersController {
	return &MembersController{service: service}
}

func (ctrl *MembersController) List(c *fiber.Ctx) error {
	claims := middlewares.Claims(c)
	members, err := ctrl.service.ListOrganizationMembers(c.Context(), claims.OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, members)
}

func (ctrl *MembersController) Delete(c *fiber.Ctx) error {
	claims := middlewares.Claims(c)
	memberID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "invalid_id", "invalid member id")
	}

	if err := ctrl.service.DeleteMember(c.Context(), claims.OrganizationID, memberID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
