package controllers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/entities"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type AuthController struct {
	cfg     *config.Config
	service *services.AuthService
}

func NewAuthController(cfg *config.Config, service *services.AuthService) *AuthController {
	return &AuthController{cfg: cfg, service: service}
}

func (ctrl *AuthController) setAuthCookies(c *fiber.Ctx, access, refresh string, accessTTL, refreshTTL time.Duration) {
	c.Cookie(&fiber.Cookie{
		Name:     "ic_access",
		Value:    access,
		HTTPOnly: true,
		Secure:   ctrl.cfg.Env == "production",
		SameSite: "Lax",
		MaxAge:   int(accessTTL.Seconds()),
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "ic_refresh",
		Value:    refresh,
		HTTPOnly: true,
		Secure:   ctrl.cfg.Env == "production",
		SameSite: "Lax",
		MaxAge:   int(refreshTTL.Seconds()),
		Path:     "/",
	})
}

func (ctrl *AuthController) clearAuthCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "ic_access",
		Value:    "",
		HTTPOnly: true,
		Secure:   ctrl.cfg.Env == "production",
		SameSite: "Lax",
		MaxAge:   -1,
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "ic_refresh",
		Value:    "",
		HTTPOnly: true,
		Secure:   ctrl.cfg.Env == "production",
		SameSite: "Lax",
		MaxAge:   -1,
		Path:     "/",
	})
}

func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req entities.RegisterRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := ctrl.service.Register(c.Context(), req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (ctrl *AuthController) SendOTP(c *fiber.Ctx) error {
	var req entities.SendOTPRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.service.SendOTP(c.Context(), req.Email); err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, entities.MessageResponse{Message: "code sent"})
}

func (ctrl *AuthController) VerifyOTP(c *fiber.Ctx) error {
	var req entities.VerifyOTPRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	access, refresh, accessTTL, refreshTTL, err := ctrl.service.VerifyOTP(c.Context(), req.Email, req.Code)
	if err != nil {
		return err
	}
	ctrl.setAuthCookies(c, access, refresh, accessTTL, refreshTTL)
	return utils.OK(c, fiber.StatusOK, entities.MessageResponse{Message: "authenticated"})
}

func (ctrl *AuthController) VerifyTOTP(c *fiber.Ctx) error {
	var req entities.VerifyTOTPRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	access, refresh, accessTTL, refreshTTL, err := ctrl.service.VerifyTOTP(c.Context(), req.Email, req.Code)
	if err != nil {
		return err
	}
	ctrl.setAuthCookies(c, access, refresh, accessTTL, refreshTTL)
	return utils.OK(c, fiber.StatusOK, entities.MessageResponse{Message: "authenticated"})
}

func (ctrl *AuthController) CreateOrganization(c *fiber.Ctx) error {
	var req entities.CreateOrganizationRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	access, refresh, accessTTL, refreshTTL, err := ctrl.service.CreateOrganization(c.Context(), middlewares.Claims(c), req)
	if err != nil {
		return err
	}
	ctrl.setAuthCookies(c, access, refresh, accessTTL, refreshTTL)
	return utils.OK(c, fiber.StatusCreated, entities.MessageResponse{Message: "organization created"})
}

// Refresh is largely obsolete via manual client calls now that the middleware
// auto-refreshes tokens, but we can keep it for completeness.
func (ctrl *AuthController) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("ic_refresh")
	if refreshToken == "" {
		return utils.ErrUnauthorized("no refresh token")
	}
	access, newRefresh, accessTTL, refreshTTL, err := ctrl.service.Refresh(c.Context(), refreshToken)
	if err != nil {
		ctrl.clearAuthCookies(c)
		return err
	}
	ctrl.setAuthCookies(c, access, newRefresh, accessTTL, refreshTTL)
	return utils.OK(c, fiber.StatusOK, entities.MessageResponse{Message: "token refreshed"})
}

func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("ic_refresh")
	if refreshToken != "" {
		_ = ctrl.service.Logout(c.Context(), refreshToken)
	}
	ctrl.clearAuthCookies(c)
	return utils.OK(c, fiber.StatusOK, entities.MessageResponse{Message: "logged out"})
}

func (ctrl *AuthController) Me(c *fiber.Ctx) error {
	resp, err := ctrl.service.Me(c.Context(), middlewares.Claims(c))
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (ctrl *AuthController) GoogleLogin(c *fiber.Ctx) error {
	intent := c.Query("intent", "login")

	url, err := ctrl.service.BuildGoogleAuthURL(intent)
	if err != nil {
		return c.Redirect(ctrl.oauthErrorRedirect(intent, err), fiber.StatusFound)
	}
	return c.Redirect(url, fiber.StatusFound)
}

// GoogleCallback now exchanges the code immediately and issues cookies, redirecting
// directly to the frontend dashboard.
func (ctrl *AuthController) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	exchangeCode, intent, err := ctrl.service.HandleGoogleCallback(c.Context(), code, state)
	if err != nil {
		return c.Redirect(ctrl.oauthErrorRedirect(intent, err), fiber.StatusFound)
	}
	
	// Trade the exchange code for the token pair directly here.
	access, refresh, accessTTL, refreshTTL, err := ctrl.service.ExchangeOAuthCode(c.Context(), exchangeCode)
	if err != nil {
		return c.Redirect(ctrl.oauthErrorRedirect(intent, err), fiber.StatusFound)
	}

	ctrl.setAuthCookies(c, access, refresh, accessTTL, refreshTTL)
	
	// Redirect back to frontend
	if intent == "register" {
		return c.Redirect(ctrl.cfg.WebOrigin+"/onboarding", fiber.StatusFound)
	}
	return c.Redirect(ctrl.cfg.WebOrigin+"/dashboard", fiber.StatusFound)
}

func (ctrl *AuthController) oauthErrorRedirect(intent string, err error) string {
	code := "oauth_failed"
	if apiErr, ok := errors.AsType[*utils.APIError](err); ok {
		code = apiErr.Code
	}
	path := "/login"
	if intent == "register" {
		path = "/register"
	}
	return ctrl.cfg.WebOrigin + path + "?error=" + code
}
