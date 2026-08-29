package devices

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
    "identitycard-server/internal/utils"
)

func (a *App) RegisterRoutes(router fiber.Router) {
    manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleMember)

    g := router.Group("/devices", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
    g.Post("/", a.handleCreate)
    g.Get("/", a.handleList)
    g.Post("/:id/revoke", a.handleRevoke)
}

func (a *App) RegisterPublicRoutes(router fiber.Router) {
    router.Post("/public/devices/pair", a.handlePair)
}

func (a *App) RegisterScannerRoutes(router fiber.Router) fiber.Router {
    g := router.Group("/scanner", a.RequireDevice())
    g.Get("/me", a.handleMe)
    return g
}

func (a *App) handleCreate(c *fiber.Ctx) error {
    var req CreateDeviceRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Create(c.Context(), middlewares.Claims(c).OrganizationID, req.Name)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusCreated, resp)
}

func (a *App) handleList(c *fiber.Ctx) error {
    resp, err := a.List(c.Context(), middlewares.Claims(c).OrganizationID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleRevoke(c *fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid device id")
    }
    if err := a.Revoke(c.Context(), middlewares.Claims(c).OrganizationID, id); err != nil {
        return err
    }
    return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handlePair(c *fiber.Ctx) error {
    var req PairDeviceRequest
    if err := utils.BindAndValidate(c, &req); err != nil {
        return err
    }
    resp, err := a.Pair(c.Context(), req.OTPCode)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleMe(c *fiber.Ctx) error {
    claims := GetDeviceClaims(c)
    resp, err := a.Me(c.Context(), claims.OrganizationID, claims.DeviceID)
    if err != nil {
        return err
    }
    return utils.OK(c, fiber.StatusOK, resp)
}
