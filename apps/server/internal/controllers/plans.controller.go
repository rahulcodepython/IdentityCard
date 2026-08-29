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

// List returns the public plan catalog — no auth required, a prospective
// customer needs to see plans before creating an account.
func (ctrl *PlansController) List(c *fiber.Ctx) error {
    resp, err := ctrl.service.List(c.Context())
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

// ListBilling returns the org's full billing picture (lineages + credits +
// available-by-type count).
func (ctrl *PlansController) ListBilling(c *fiber.Ctx) error {
    resp, err := ctrl.service.ListBilling(c.Context(), middlewares.Claims(c).OrganizationID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

// Purchase starts a new billing lineage for the org — always additive, never
// replaces an existing one (see PlansService.Purchase).
func (ctrl *PlansController) Purchase(c *fiber.Ctx) error {
    var req entities.PurchaseRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := ctrl.service.Purchase(c.Context(), middlewares.Claims(c).OrganizationID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusCreated, resp)
}

// Renew records payment for the current period of a billing lineage and
// spawns the next one. For flash lineages this also mints one new flash
// credit (the literal "gate" for getting one more flash event).
func (ctrl *PlansController) Renew(c *fiber.Ctx) error {
    lineageRootID, err := parseLineageRootID(c)
    if err != nil {
        return err
    }
    resp, err := ctrl.service.Renew(c.Context(), middlewares.Claims(c).OrganizationID, lineageRootID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

// Upgrade switches a billing lineage's current period to a new plan in place.
// No new credits are minted immediately; the new plan's credit terms apply
// starting the next renewal.
func (ctrl *PlansController) Upgrade(c *fiber.Ctx) error {
    lineageRootID, err := parseLineageRootID(c)
    if err != nil {
        return err
    }
    var req entities.UpgradeRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := ctrl.service.Upgrade(c.Context(), middlewares.Claims(c).OrganizationID, lineageRootID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

// Cancel marks a billing lineage's current period 'cancel' — the next daily
// sync restricts every credit funded by that lineage.
func (ctrl *PlansController) Cancel(c *fiber.Ctx) error {
    lineageRootID, err := parseLineageRootID(c)
    if err != nil {
        return err
    }
    resp, err := ctrl.service.Cancel(c.Context(), middlewares.Claims(c).OrganizationID, lineageRootID)
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
