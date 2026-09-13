package middlewares

import (
    "strings"
    "uuid"

    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/jwt"
    "identitycard-server/internal/utils"
)

// ExtractToken retrieves the authentication token from either the Authorization header
// (with or without 'Bearer' prefix) or from session cookies.
func ExtractToken(c *fiber.Ctx) string {
    authHeader := strings.TrimSpace(c.Get(generic.HeaderAuthorization))
    if authHeader != "" {
        if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
            return strings.TrimSpace(authHeader[7:])
        }
        if !strings.Contains(authHeader, " ") {
            return authHeader
        }
    }

    // Fallback: check session cookies
    if cookieToken := strings.TrimSpace(c.Cookies("token")); cookieToken != "" {
        return cookieToken
    }
    if sessionCookie := strings.TrimSpace(c.Cookies("better-auth.session_token")); sessionCookie != "" {
        return sessionCookie
    }

    return ""
}

// BaseAuthMiddleware verifies the bearer JWT issued by better-auth and populates
// the authenticated claims on the Fiber request context. verifier is
// constructed once at startup (cmd/server/main.go) since it holds a
// background refresh goroutine for the JWKS keyset, and passed in here like
// every other dependency rather than held as package-level state.
func BaseAuthMiddleware(verifier *jwt.Verifier) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := ExtractToken(c)
        if token == "" {
            return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
        }

        if verifier == nil {
            return utils.ErrInternal("Auth verifier not initialized", nil)
        }

        claims, err := verifier.Parse(token)
        if err != nil {
            return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
        }

        userUUID, err := claims.UserID()
        if err != nil || userUUID == uuid.Nil() {
            return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
        }

        c.Locals(generic.ContextKeyClaims, claims)
        return c.Next()
    }
}

// Claims returns the authenticated request's claims from Fiber locals.
// Returns nil if unauthenticated.
func Claims(c *fiber.Ctx) *jwt.Claims {
    claims, _ := c.Locals(generic.ContextKeyClaims).(*jwt.Claims)
    return claims
}
