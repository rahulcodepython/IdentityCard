package middlewares

import (
    "slices"
    "strings"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/jwt"
    "identitycard-server/internal/utils"
)

// BaseAuthMiddleware verifies the bearer JWT issued by better-auth and populates
// the authenticated claims on the Fiber request context. verifier is
// constructed once at startup (cmd/server/main.go) since it holds a
// background refresh goroutine for the JWKS keyset, and passed in here like
// every other dependency rather than held as package-level state.
func BaseAuthMiddleware(verifier *jwt.Verifier) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get(generic.HeaderAuthorization)
        token, ok := strings.CutPrefix(authHeader, "Bearer ")
        if !ok || token == "" {
            return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
        }

        if verifier == nil {
            return utils.ErrInternal("Auth verifier not initialized", nil)
        }

        claims, err := verifier.Parse(token)
        if err != nil {
            return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
        }

        if claims.MustChangePassword {
            return utils.ErrUnauthorized("Password change required before API access.")
        }

        if claims.Banned {
            return utils.ErrUnauthorized("User account is banned.")
        }

        c.Locals(generic.ContextKeyClaims, claims)
        return c.Next()
    }
}

// NewAuthMiddleware is an alias for BaseAuthMiddleware.
func NewAuthMiddleware(v *jwt.Verifier) fiber.Handler {
    return BaseAuthMiddleware(v)
}

// Claims returns the authenticated request's claims from Fiber locals.
// Returns nil if unauthenticated.
func Claims(c *fiber.Ctx) *jwt.Claims {
    claims, _ := c.Locals(generic.ContextKeyClaims).(*jwt.Claims)
    return claims
}

// RequireRole grants access if the caller holds ANY of the given roles.
// Must run after BaseAuthMiddleware.
func RequireRole(roles ...generic.Role) fiber.Handler {
    return func(c *fiber.Ctx) error {
        claims := Claims(c)
        if claims == nil {
            return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
        }
        if slices.ContainsFunc(roles, claims.HasRole) {
            return c.Next()
        }
        return utils.ErrForbidden(generic.ErrMsgForbidden, nil)
    }
}

// RequireOrganization requires the caller's access token to carry an organization.
// Must run after BaseAuthMiddleware.
func RequireOrganization(c *fiber.Ctx) error {
    claims := Claims(c)
    if claims == nil || claims.OrganizationID == uuid.Nil {
        return utils.ErrForbidden(generic.ErrMsgOrganizationRequired, nil)
    }
    return c.Next()
}

