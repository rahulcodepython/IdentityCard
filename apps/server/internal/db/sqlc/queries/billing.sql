-- name: CreateBillingLineageRoot :one
-- Starts a new billing lineage — inserts the first period row, then
-- points its own lineage_root_id at itself in the same statement, so
-- "current status of this lineage" is always a lookup by
-- lineage_root_id, uniformly for the root row and every renewal after it.
WITH new_billing AS (
    INSERT INTO billing (organization_id, plan_id, billing_number, period_start, period_end, status, amount)
    VALUES ($1, $2, 1, $3, $4, $5, $6)
    RETURNING id, organization_id, plan_id, billing_number, period_start, period_end, status, amount, paid_at, transaction_id, created_at
)
UPDATE billing SET lineage_root_id = new_billing.id
FROM new_billing
WHERE billing.id = new_billing.id
RETURNING billing.*;

-- name: CreateBillingRenewal :one
-- The next period row in an existing lineage — spawned once the previous
-- row's payment lands (see MarkBillingPaid).
INSERT INTO billing (organization_id, plan_id, lineage_root_id, billing_number, period_start, period_end, status, amount)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetBilling :one
SELECT * FROM billing WHERE id = $1 AND organization_id = $2;

-- name: GetLatestBillingForLineage :one
SELECT * FROM billing WHERE lineage_root_id = $1 ORDER BY period_start DESC LIMIT 1;

-- name: ListLatestBillingForOrganization :many
-- One row per lineage — the org's whole billing picture (see
-- PlansService.ListBilling).
SELECT DISTINCT ON (lineage_root_id) *
FROM billing
WHERE organization_id = $1
ORDER BY lineage_root_id, period_start DESC;

-- name: MarkBillingPaid :one
UPDATE billing SET status = 'paid', paid_at = now(), transaction_id = $2
WHERE id = $1
RETURNING *;

-- name: CancelBillingLineage :one
UPDATE billing SET status = 'cancel'
WHERE id = (SELECT b.id FROM billing b WHERE b.lineage_root_id = $1 ORDER BY b.period_start DESC LIMIT 1)
RETURNING *;

-- name: UpgradeBillingPlan :one
-- Upgrade updates the current (latest) period row's plan/amount in
-- place — no new transaction/billing/credits rows spawned immediately;
-- the new plan's terms apply starting the next renewal.
UPDATE billing SET plan_id = $2, amount = $3
WHERE id = (SELECT b.id FROM billing b WHERE b.lineage_root_id = $1 ORDER BY b.period_start DESC LIMIT 1)
RETURNING *;

-- name: FlipLapsedBillingToPending :many
-- The daily cron's first step: every lineage's latest row, if it's
-- active and its due date has passed, flips to pending. Targets only the
-- latest row per lineage (a lineage's historical paid rows are untouched).
UPDATE billing b
SET status = 'pending'
WHERE b.status = 'active' AND b.period_end < $1::date
  AND b.id = (SELECT id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
RETURNING *;
