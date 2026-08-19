package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/middleware"
)

type Handler struct {
	cfg     *config.Config
	service *Service
}

func NewHandler(cfg *config.Config, service *Service) *Handler {
	return &Handler{cfg: cfg, service: service}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Register(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusCreated, resp)
}

func (h *Handler) SendOTP(c *fiber.Ctx) error {
	var req SendOTPRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.service.SendOTP(c.Context(), req.Email); err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, MessageResponse{Message: "code sent"})
}

func (h *Handler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	access, refresh, accessTTL, refreshTTL, err := h.service.VerifyOTP(c.Context(), req.Email, req.Code)
	if err != nil {
		return err
	}
	coreauth.SetAuthCookies(c, h.cfg, access, refresh, accessTTL, refreshTTL)
	return httpx.OK(c, fiber.StatusOK, MessageResponse{Message: "verified"})
}

func (h *Handler) VerifyTOTP(c *fiber.Ctx) error {
	var req VerifyTOTPRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	access, refresh, accessTTL, refreshTTL, err := h.service.VerifyTOTP(c.Context(), req.Email, req.Code)
	if err != nil {
		return err
	}
	coreauth.SetAuthCookies(c, h.cfg, access, refresh, accessTTL, refreshTTL)
	return httpx.OK(c, fiber.StatusOK, MessageResponse{Message: "verified"})
}

func (h *Handler) CreateOrganization(c *fiber.Ctx) error {
	var req CreateOrganizationRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	access, refresh, accessTTL, refreshTTL, err := h.service.CreateOrganization(c.Context(), middleware.Claims(c), req)
	if err != nil {
		return err
	}
	coreauth.SetAuthCookies(c, h.cfg, access, refresh, accessTTL, refreshTTL)
	return httpx.OK(c, fiber.StatusCreated, MessageResponse{Message: "organization created"})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(coreauth.RefreshCookieName)
	if refreshToken == "" {
		return httpx.ErrUnauthorized("")
	}

	access, newRefresh, accessTTL, refreshTTL, err := h.service.Refresh(c.Context(), refreshToken)
	if err != nil {
		coreauth.ClearAuthCookies(c, h.cfg)
		return err
	}

	coreauth.SetAuthCookies(c, h.cfg, access, newRefresh, accessTTL, refreshTTL)
	return httpx.OK(c, fiber.StatusOK, MessageResponse{Message: "refreshed"})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	if refreshToken := c.Cookies(coreauth.RefreshCookieName); refreshToken != "" {
		_ = h.service.Logout(c.Context(), refreshToken)
	}
	coreauth.ClearAuthCookies(c, h.cfg)
	return httpx.OK(c, fiber.StatusOK, MessageResponse{Message: "logged out"})
}

func (h *Handler) Me(c *fiber.Ctx) error {
	resp, err := h.service.Me(c.Context(), middleware.Claims(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

// GoogleLogin and GoogleCallback are full-page browser navigations, not
// fetch calls — every outcome (success or error) is a redirect, never the
// standard {"error": ...} JSON envelope, since that would render as raw
// JSON in the browser mid-flow.
func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	intent := c.Query("intent", "login")

	url, err := h.service.BuildGoogleAuthURL(intent)
	if err != nil {
		return c.Redirect(h.oauthErrorRedirect(intent, err), fiber.StatusFound)
	}
	return c.Redirect(url, fiber.StatusFound)
}

func (h *Handler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	access, refresh, redirectPath, intent, accessTTL, refreshTTL, err := h.service.HandleGoogleCallback(c.Context(), code, state)
	if err != nil {
		return c.Redirect(h.oauthErrorRedirect(intent, err), fiber.StatusFound)
	}

	coreauth.SetAuthCookies(c, h.cfg, access, refresh, accessTTL, refreshTTL)
	return c.Redirect(h.cfg.WebOrigin+redirectPath, fiber.StatusFound)
}

func (h *Handler) oauthErrorRedirect(intent string, err error) string {
	code := "oauth_failed"
	if apiErr, ok := errors.AsType[*httpx.APIError](err); ok {
		code = apiErr.Code
	}
	path := "/login"
	if intent == "register" {
		path = "/register"
	}
	return h.cfg.WebOrigin + path + "?error=" + code
}
