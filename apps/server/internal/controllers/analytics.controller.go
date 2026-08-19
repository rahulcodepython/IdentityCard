package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type AnalyticsController struct {
	service *services.AnalyticsService
}

func NewAnalyticsController(service *services.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{service: service}
}

func (ctrl *AnalyticsController) Summary(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	resp, err := ctrl.service.Summary(c.Context(), middlewares.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *AnalyticsController) Overview(c *fiber.Ctx) error {
	resp, err := ctrl.service.Overview(c.Context(), middlewares.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *AnalyticsController) Daily(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}

	var subEventID *uuid.UUID
	if raw := c.Query("sub_event_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub_event_id")
		}
		subEventID = &id
	}

	resp, err := ctrl.service.Daily(c.Context(), middlewares.Claims(c).OrganizationID, eventID, subEventID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}
