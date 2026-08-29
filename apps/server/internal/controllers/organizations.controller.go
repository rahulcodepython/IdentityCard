package controllers

import (
	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type OrganizationsController struct {
	service *services.OrganizationsService
}

func NewOrganizationsController(service *services.OrganizationsService) *OrganizationsController {
	return &OrganizationsController{service: service}
}

func (ctrl *OrganizationsController) GetSettings(c *fiber.Ctx) error {
	claims := middlewares.Claims(c)
	resp, err := ctrl.service.GetSettings(c.Context(), claims.OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *OrganizationsController) UpdateSettings(c *fiber.Ctx) error {
	var req entities.UpdateOrganizationSettingsRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	claims := middlewares.Claims(c)
	resp, err := ctrl.service.UpdateSettings(c.Context(), claims.OrganizationID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *OrganizationsController) DeleteOrganization(c *fiber.Ctx) error {
	claims := middlewares.Claims(c)
	if err := ctrl.service.DeleteOrganization(c.Context(), claims.OrganizationID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *OrganizationsController) DeleteLogo(c *fiber.Ctx) error {
	claims := middlewares.Claims(c)
	if err := ctrl.service.DeleteLogo(c.Context(), claims.OrganizationID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *OrganizationsController) UploadLogo(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("logo")
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "missing logo file")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return utils.ErrInternal()
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	claims := middlewares.Claims(c)
	if err := ctrl.service.UploadLogo(c.Context(), claims.OrganizationID, contentType, fileHeader.Size, file); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetLogo streams the image directly — like the people CSV export, a
// deliberate exception to the {"data": ...} envelope for a binary response.
func (ctrl *OrganizationsController) GetLogo(c *fiber.Ctx) error {
	claims := middlewares.Claims(c)
	data, contentType, err := ctrl.service.GetLogo(c.Context(), claims.OrganizationID)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, contentType)
	return c.Send(data)
}
