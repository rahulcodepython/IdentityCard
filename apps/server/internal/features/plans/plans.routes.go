package plans

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/utils"
)

func (a *App) RegisterRoutes(router fiber.Router) {
	router.Get("/plans", a.handleListPlans)

	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)
	g := router.Group("/plans", middlewares.RequireAuth, middlewares.RequireOrganization)
	g.Get("/billing", a.handleListBilling)
	g.Post("/purchase", manage, a.handlePurchase)
	g.Post("/billing/:lineageRootId/renew", manage, a.handleRenew)
	g.Post("/billing/:lineageRootId/upgrade", manage, a.handleUpgrade)
	g.Post("/billing/:lineageRootId/cancel", manage, a.handleCancel)
}

func (a *App) handleListPlans(c *fiber.Ctx) error {
	resp, err := a.List(c.Context())
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleListBilling(c *fiber.Ctx) error {
	resp, err := a.ListBilling(c.Context(), middlewares.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handlePurchase(c *fiber.Ctx) error {
	var req PurchaseRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := a.Purchase(c.Context(), middlewares.Claims(c).OrganizationID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (a *App) handleRenew(c *fiber.Ctx) error {
	lineageRootID, err := parseLineageRootID(c)
	if err != nil {
		return err
	}
	resp, err := a.Renew(c.Context(), middlewares.Claims(c).OrganizationID, lineageRootID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpgrade(c *fiber.Ctx) error {
	lineageRootID, err := parseLineageRootID(c)
	if err != nil {
		return err
	}
	var req UpgradeRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := a.Upgrade(c.Context(), middlewares.Claims(c).OrganizationID, lineageRootID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleCancel(c *fiber.Ctx) error {
	lineageRootID, err := parseLineageRootID(c)
	if err != nil {
		return err
	}
	resp, err := a.Cancel(c.Context(), middlewares.Claims(c).OrganizationID, lineageRootID)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func parseLineageRootID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("lineageRootId"))
	if err != nil {
		return uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid lineage root id")
	}
	return id, nil
}
