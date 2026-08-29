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

const claimsLocalsKey = "auth_claims"

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
// read. Go is a pure resource server now — no cookies, no refresh-token
// dance; better-auth owns session/refresh entirely and this only ever
// sees short-lived access tokens.
func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get(fiber.HeaderAuthorization)
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || token == "" {
		return utils.ErrUnauthorized("missing authentication")
	}

	claims, err := verifier.Parse(token)
	if err != nil {
		return utils.ErrUnauthorized("invalid or expired token")
	}

	c.Locals(claimsLocalsKey, claims)
	return c.Next()
}

// Claims returns the authenticated request's claims. It must only be
// called after RequireAuth has run for the route.
func Claims(c *fiber.Ctx) *jwt.Claims {
	claims, _ := c.Locals(claimsLocalsKey).(*jwt.Claims)
	return claims
}

// RequireRole grants access if the caller holds ANY of the given roles.
// Must run after RequireAuth.
func RequireRole(roles ...generic.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := Claims(c)
		if claims == nil {
			return utils.ErrUnauthorized("")
		}
		if slices.ContainsFunc(roles, claims.HasRole) {
			return c.Next()
		}
		return utils.ErrForbidden("")
	}
}

// RequireOrganization additionally requires the caller's access token to
// carry an organization. A signed-in user with no organization yet (no
// purchase/onboarding completed — see the organization plugin's
// creation gating in apps/web/lib/auth.ts) has none, and every business
// route needs one to scope its queries by. Must run after RequireAuth.
func RequireOrganization(c *fiber.Ctx) error {
	claims := Claims(c)
	if claims == nil || claims.OrganizationID == uuid.Nil {
		return utils.NewError(fiber.StatusForbidden, "organization_required", "finish onboarding first")
	}
	return c.Next()
}
