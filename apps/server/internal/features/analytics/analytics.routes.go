package analytics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/utils"
)

func (a *App) RegisterRoutes(router fiber.Router) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

	g := router.Group("/events/:eventId/analytics", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
	g.Get("/summary", a.handleSummary)
	g.Get("/daily", a.handleDaily)

	router.Get("/analytics/overview", middlewares.RequireAuth, middlewares.RequireOrganization, manage, a.handleOverview)
}

func (a *App) handleSummary(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	resp, err := a.Summary(c.Context(), middlewares.Claims(c).OrganizationID, eventID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleOverview(c *fiber.Ctx) error {
	resp, err := a.Overview(c.Context(), middlewares.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDaily(c *fiber.Ctx) error {
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

	resp, err := a.Daily(c.Context(), middlewares.Claims(c).OrganizationID, eventID, subEventID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}
