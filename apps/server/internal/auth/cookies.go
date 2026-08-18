package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
)

const (
	AccessCookieName  = "ic_access"
	RefreshCookieName = "ic_refresh"
)

// SetAuthCookies writes both tokens as httpOnly cookies so they never
// reach client-side JS. RefreshCookieName is scoped to the auth refresh
// path only, so a stolen access-token cookie can't be used to mint new
// sessions.
func SetAuthCookies(c *fiber.Ctx, cfg *config.Config, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) {
	c.Cookie(&fiber.Cookie{
		Name:     AccessCookieName,
		Value:    accessToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		Expires:  time.Now().Add(accessTTL),
		HTTPOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
	c.Cookie(&fiber.Cookie{
		Name:     RefreshCookieName,
		Value:    refreshToken,
		Path:     "/auth",
		Domain:   cfg.CookieDomain,
		Expires:  time.Now().Add(refreshTTL),
		HTTPOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func ClearAuthCookies(c *fiber.Ctx, cfg *config.Config) {
	c.Cookie(&fiber.Cookie{
		Name: AccessCookieName, Value: "", Path: "/", Domain: cfg.CookieDomain,
		Expires: time.Now().Add(-time.Hour), HTTPOnly: true, Secure: cfg.CookieSecure,
	})
	c.Cookie(&fiber.Cookie{
		Name: RefreshCookieName, Value: "", Path: "/auth", Domain: cfg.CookieDomain,
		Expires: time.Now().Add(-time.Hour), HTTPOnly: true, Secure: cfg.CookieSecure,
	})
}
