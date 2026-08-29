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

var verifier *jwt.Verifier

// InitAuth wires the JWKS verifier used by RequireAuth — called once at
// startup (see cmd/server/main.go) since it holds a background refresh
// goroutine for the keyset.
func InitAuth(v *jwt.Verifier) {
    verifier = v
}

// RequireAuth verifies the bearer JWT better-auth issued (see
// apps/web/lib/auth.ts's jwt plugin) and, on success, stores the parsed
// claims on the request context for downstream handlers/middleware to
// read.
func RequireAuth(c *fiber.Ctx) error {
    authHeader := c.Get(generic.HeaderAuthorization)
    token, ok := strings.CutPrefix(authHeader, "Bearer ")
    if !ok || token == "" {
        return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
    }

    claims, err := verifier.Parse(token)
    if err != nil {
        return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
    }

    c.Locals(generic.ContextKeyClaims, claims)
    return c.Next()
}

// Claims returns the authenticated request's claims. It must only be
// called after RequireAuth has run for the route.
func Claims(c *fiber.Ctx) *jwt.Claims {
    claims, _ := c.Locals(generic.ContextKeyClaims).(*jwt.Claims)
    return claims
}

// RequireRole grants access if the caller holds ANY of the given roles.
// Must run after RequireAuth.
func RequireRole(roles ...generic.Role) fiber.Handler {
    return func(c *fiber.Ctx) error {
        claims := Claims(c)
        if claims == nil {
            return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
        }
        if slices.ContainsFunc(roles, claims.HasRole) {
            return c.Next()
        }
        return utils.ErrForbidden(generic.ErrMsgForbidden)
    }
}

// RequireOrganization additionally requires the caller's access token to
// carry an organization. Must run after RequireAuth.
func RequireOrganization(c *fiber.Ctx) error {
    claims := Claims(c)
    if claims == nil || claims.OrganizationID == uuid.Nil {
        return utils.NewError(fiber.StatusForbidden, "organization_required", generic.ErrMsgOrganizationRequired)
    }
    return c.Next()
}
