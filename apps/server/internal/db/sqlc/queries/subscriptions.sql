-- name: CreateSubscription :one
-- Always additive — see plans.Service.Subscribe. Never cancels or
-- replaces an existing subscription; an org can hold many concurrently
-- (that's how a Custom top-up stacks on top of a Base/Custom/Unlimited
-- plan instead of replacing it).
INSERT INTO subscriptions (organization_id, plan_id, kind, billing_cycle, event_quota, current_period_end)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CreateSubscriptionPeriod :one
INSERT INTO subscription_periods (subscription_id, period_start, period_end)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListSubscriptionsForOrganization :many
SELECT * FROM subscriptions WHERE organization_id = $1 ORDER BY started_at DESC;

-- name: ListActiveSubscriptionsForOrganization :many
-- Oldest-first — see events.Service.checkEventLimit, which spends an
-- org's earliest-purchased capacity before newer top-ups, so a
-- subscription that later lapses only ever "owns" the events it actually
-- funded.
SELECT * FROM subscriptions WHERE organization_id = $1 AND status = 'active' ORDER BY started_at ASC;

-- name: GetSubscriptionForOrganization :one
SELECT * FROM subscriptions WHERE id = $1 AND organization_id = $2;

-- name: RenewSubscription :one
UPDATE subscriptions
SET status = 'active', current_period_end = $2, past_due_since = NULL, grace_deadline = NULL, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListPastDueCandidateSubscriptions :many
-- Recurring subscriptions whose period lapsed without a renewal. Flash
-- (billing_cycle='one_time') is excluded on purpose — it has its own
-- retention rule (see ListActiveFlashSubscriptions) and never goes
-- past_due.
SELECT * FROM subscriptions
WHERE status = 'active' AND billing_cycle IN ('monthly', 'yearly') AND current_period_end < $1;

-- name: MarkSubscriptionPastDue :exec
UPDATE subscriptions SET status = 'past_due', past_due_since = $2, grace_deadline = $3, updated_at = now()
WHERE id = $1;

-- name: ListGraceExpiredSubscriptions :many
SELECT * FROM subscriptions WHERE status = 'past_due' AND grace_deadline < $1;

-- name: MarkSubscriptionExpired :exec
UPDATE subscriptions SET status = 'expired', updated_at = now() WHERE id = $1;

-- name: ListActiveFlashSubscriptions :many
SELECT * FROM subscriptions WHERE kind = 'flash' AND status = 'active';
