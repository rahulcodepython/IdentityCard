package subevents

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/utils"
)

func (a *App) handleCreate(c *fiber.Ctx) error {
	eventID, _, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	var req CreateSubEventRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := a.Create(c.Context(), middlewares.Claims(c).OrganizationID, eventID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (a *App) handleList(c *fiber.Ctx) error {
	eventID, _, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	resp, err := a.List(c.Context(), middlewares.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleGet(c *fiber.Ctx) error {
	eventID, id, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	resp, err := a.Get(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpdate(c *fiber.Ctx) error {
	eventID, id, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	var req UpdateSubEventRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := a.Update(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDelete(c *fiber.Ctx) error {
	eventID, id, err := parseSubEventIDParams(c)
	if err != nil {
		return err
	}
	if err := a.Delete(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
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
