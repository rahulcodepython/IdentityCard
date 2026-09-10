package middlewares

import (
    "fmt"
    "log/slog"
    "math/rand/v2"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/cache"
    "identitycard-server/internal/utils"
)

const (
    BillingCachePrefix       = "org:billing_access:"
    BillingStatusActive      = "active"
    BillingStatusGracePeriod = "grace_period"
    BillingStatusPruned      = "pruned"
    BillingStatusNoBilling   = "no_billing"

    BaseBillingTTL = 5 * time.Minute
    JitterTTL      = 1 * time.Minute
)

// BillingAccessState tracks the cached subscription status for an organization.
type BillingAccessState struct {
    Status    string    `json:"status"`
    PeriodEnd time.Time `json:"period_end"`
    PruneDate time.Time `json:"prune_date"`
}

// CalculateJitteredTTL adds a random jitter in the range [-jitterMax, +jitterMax]
// to the base duration to prevent Redis cache stampedes.
func CalculateJitteredTTL(base time.Duration, jitterMax time.Duration) time.Duration {
    if jitterMax <= 0 {
        return base
    }
    jitter := time.Duration(rand.Int64N(int64(jitterMax*2))) - jitterMax
    ttl := base + jitter
    if ttl <= 0 {
        return base
    }
    return ttl
}

// RequireActiveBilling enforces subscription authorization for organization routes.
// It checks Redis first with jittered TTL, falling back to PostgreSQL.
func RequireActiveBilling(pool *pgxpool.Pool, cch *cache.Cache) fiber.Handler {
    const checkBillingQuery = `
        SELECT 
            CASE
                WHEN EXISTS (
                    SELECT 1 FROM billing 
                    WHERE organization_id = $1 
                      AND status IN ('active', 'paid') 
                      AND period_end >= CURRENT_DATE
                ) THEN 'active'
                WHEN EXISTS (
                    SELECT 1 FROM billing 
                    WHERE organization_id = $1 
                      AND prune_date >= CURRENT_DATE
                ) THEN 'grace_period'
                WHEN EXISTS (
                    SELECT 1 FROM billing 
                    WHERE organization_id = $1
                ) THEN 'pruned'
                ELSE 'no_billing'
            END AS status,
            COALESCE((
                SELECT MAX(period_end) FROM billing WHERE organization_id = $1
            ), '1970-01-01'::date) AS period_end,
            COALESCE((
                SELECT MAX(prune_date) FROM billing WHERE organization_id = $1
            ), '1970-01-01'::date) AS prune_date;
    `

    return func(c *fiber.Ctx) error {
        path := c.Path()

        // 1. Route exemptions: Organizations and Plans endpoints must always be accessible
        // so owners can inspect billing, upgrade, renew, or configure organization settings.
        if strings.HasPrefix(path, generic.APIV1Prefix+"/plans") ||
            strings.HasPrefix(path, generic.APIV1Prefix+"/organizations") {
            return c.Next()
        }

        claims := Claims(c)
        if claims == nil || claims.OrganizationID == uuid.Nil {
            // Unauthenticated or non-org requests deferred to subsequent handlers (e.g. RequireOrganization)
            return c.Next()
        }

        ctx := c.UserContext()
        orgID := claims.OrganizationID
        cacheKey := fmt.Sprintf("%s%s", BillingCachePrefix, orgID.String())

        var state BillingAccessState
        cacheHit := false

        if cch != nil {
            hit, err := cch.Get(ctx, cacheKey, &state)
            if err == nil && hit {
                cacheHit = true
            }
        }

        if !cacheHit {
            if pool == nil {
                return utils.ErrInternal("Database pool not available for billing authorization", nil)
            }

            err := pool.QueryRow(ctx, checkBillingQuery, orgID).Scan(
                &state.Status,
                &state.PeriodEnd,
                &state.PruneDate,
            )
            if err != nil {
                slog.Error("failed to query billing access status", "org_id", orgID, "error", err)
                return utils.ErrInternal("Failed to verify organization billing status", err)
            }

            if cch != nil {
                ttl := CalculateJitteredTTL(BaseBillingTTL, JitterTTL)
                _ = cch.Set(ctx, cacheKey, state, ttl)
            }
        }

        // 2. Authorization enforcement based on lifecycle state
        switch state.Status {
        case BillingStatusActive:
            return c.Next()

        case BillingStatusGracePeriod:
            c.Set("X-Billing-Status", "grace_period")
            // In grace period: Block creation or publishing operations
            isPublish := c.Method() == fiber.MethodPost && strings.HasSuffix(path, "/publish")
            isCreateEvent := c.Method() == fiber.MethodPost && (path == generic.APIV1Prefix+"/events" || path == generic.APIV1Prefix+"/events/")

            if isPublish || isCreateEvent {
                return utils.NewError(fiber.StatusPaymentRequired, "Subscription in grace period. Renew to create or publish events.", nil)
            }
            return c.Next()

        case BillingStatusPruned, BillingStatusNoBilling:
            return utils.NewError(fiber.StatusPaymentRequired, "Subscription expired. Please subscribe to a plan to continue.", nil)

        default:
            return utils.NewError(fiber.StatusPaymentRequired, "Valid billing subscription required.", nil)
        }
    }
}
