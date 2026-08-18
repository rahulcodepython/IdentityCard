package auth

import (
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

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}

	access, refresh, accessTTL, refreshTTL, err := h.service.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	coreauth.SetAuthCookies(c, h.cfg, access, refresh, accessTTL, refreshTTL)
	return httpx.OK(c, fiber.StatusOK, MessageResponse{Message: "logged in"})
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}

	access, refresh, accessTTL, refreshTTL, err := h.service.Register(c.Context(), req)
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
