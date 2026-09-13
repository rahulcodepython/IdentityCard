package billing

import (
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/utils"
)

func (a *App) handleGetBilling(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }

    resp, err := a.GetOverview(c.Context(), orgID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handlePurchaseCredits(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }

    var req PurchaseCreditsRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }

    resp, err := a.PurchaseCredits(c.Context(), orgID, req)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleRenewAnnual(c *fiber.Ctx) error {
    orgID, err := utils.ParseOrgID(c)
    if err != nil {
        return err
    }

    resp, err := a.RenewAnnual(c.Context(), orgID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}
