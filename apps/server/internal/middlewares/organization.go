package middlewares

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"identitycard-server/internal/generic"
	"identitycard-server/internal/pkg/cache"
	"identitycard-server/internal/utils"
)

// RequireOrganizationMember ensures that the caller is an active member of the
// organization specified in the route parameter :orgId, and attaches their member
// roles as a map[string]bool to request locals for downstream role checks.
func RequireOrganizationMember(pool *pgxpool.Pool, cch *cache.Cache) fiber.Handler {
	const query = `SELECT "role" FROM "member" WHERE "userId" = $1 AND "organizationId" = $2 LIMIT 1;`

	return func(c *fiber.Ctx) error {
		orgID, err := utils.ParseOrgID(c)
		if err != nil {
			return err
		}

		claims := Claims(c)
		if claims == nil {
			return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
		}

		userUUID, err := claims.UserID()
		if err != nil || userUUID == uuid.Nil() {
			return utils.ErrUnauthorized(generic.ErrMsgUnauthorized)
		}

		ctx := c.UserContext()
		cacheKey := fmt.Sprintf("%s%s:%s", generic.MemberRoleCachePrefix, userUUID.String(), orgID.String())

		entry, err := cache.FetchOrCompute(ctx, cch, cache.FetchOptions[string]{
			Key:     cacheKey,
			TTL:     5 * time.Minute,
			IsValid: func(r string) bool { return r != "" },
			Fetch: func(fetchCtx context.Context) (string, error) {
				if pool == nil {
					return "", utils.ErrInternal("Database pool not available for organization authorization", nil)
				}

				var role string
				if err := pool.QueryRow(fetchCtx, query, userUUID, orgID).Scan(&role); err != nil {
					return "", utils.ErrForbidden(generic.ErrMsgForbidden, err)
				}
				return role, nil
			},
			OnHit: func(_ context.Context, _ string, _ string) {
				// preferred action on cache hit can be registered here
			},
			OnMiss: func(_ context.Context, _ string, _ string) {
				// preferred action on cache miss can be registered here
			},
		})
		if err != nil {
			return err
		}

		// Store roles as map format in c.Locals as requested
		rolesMap := map[string]bool{
			entry.Data: true,
		}
		c.Locals(generic.ContextKeyRoles, rolesMap)
		return c.Next()
	}
}

// RequireRole returns a handler that enforces the caller's organization role is in the allowed set.
func RequireRole(roles ...generic.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoles, _ := c.Locals(generic.ContextKeyRoles).(map[string]bool)
		if userRoles == nil {
			return utils.ErrForbidden(generic.ErrMsgForbidden, nil)
		}
		for _, r := range roles {
			if userRoles[string(r)] {
				return c.Next()
			}
		}
		return utils.ErrForbidden(generic.ErrMsgForbidden, nil)
	}
}

// Roles returns the caller's organization roles map from Fiber locals.
func Roles(c *fiber.Ctx) map[string]bool {
	roles, _ := c.Locals(generic.ContextKeyRoles).(map[string]bool)
	return roles
}

// HasRole checks if the caller holds a specific role in the active organization.
func HasRole(c *fiber.Ctx, role generic.Role) bool {
	roles := Roles(c)
	return roles != nil && roles[string(role)]
}
