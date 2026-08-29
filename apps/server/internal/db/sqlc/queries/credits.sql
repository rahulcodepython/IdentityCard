-- name: CreateCredit :one
INSERT INTO credits (organization_id, billing_id, type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: FindAvailableCredit :one
-- FIFO, unconsumed, non-restricted credit matching one of the acceptable
-- types for the event being created (e.g. {'flash','all'} for a flash
-- event, {'events','all'} for standard/grouped) — one query covers both
-- the required type and the 'all' (unlimited-plan) fallback while
-- keeping true creation-order FIFO across them. A flash-type credit
-- additionally must have been granted today — this is how a flash
-- credit "expires that very day", no separate expiry sweep needed.
-- FOR UPDATE so two concurrent creates can't both consume the same row.
SELECT * FROM credits
WHERE organization_id = $1
  AND type = ANY(sqlc.arg('types')::text[])
  AND event_id IS NULL
  AND is_restricted = false
  AND (type <> 'flash' OR created_at::date = CURRENT_DATE)
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE;

-- name: ConsumeCredit :one
UPDATE credits SET event_id = $2 WHERE id = $1 RETURNING *;

-- name: GetCreditForEvent :one
SELECT * FROM credits WHERE event_id = $1;

-- name: GetActiveUnlimitedLineageRoot :one
-- The org's active/paid unlimited-plan billing lineage, if it has one —
-- used to decide whether to replenish an 'all' credit immediately after
-- one is consumed (see PlansService's unlimited-replenishment rule).
SELECT b.lineage_root_id FROM billing b
JOIN plans p ON p.id = b.plan_id
WHERE b.organization_id = $1 AND p.kind = 'unlimited'
  AND b.status IN ('active', 'paid')
  AND b.id = (SELECT id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
LIMIT 1;

-- name: SyncCreditRestriction :many
-- The daily cron's second step: recompute every credit's is_restricted
-- from its lineage's current latest billing status, only touching rows
-- whose value actually changes (so this is a no-op most days for most
-- credits) — pending/cancel restricts, active/paid clears.
-- restricted_since is stamped on the transition into restricted and
-- cleared on the transition out.
UPDATE credits c
SET is_restricted = (latest.status IN ('pending', 'cancel')),
    restricted_since = CASE
        WHEN latest.status IN ('pending', 'cancel') AND NOT c.is_restricted THEN now()
        WHEN latest.status NOT IN ('pending', 'cancel') THEN NULL
        ELSE c.restricted_since
    END
FROM (
    SELECT DISTINCT ON (lineage_root_id) lineage_root_id, status
    FROM billing
    ORDER BY lineage_root_id, period_start DESC
) latest
WHERE c.billing_id = latest.lineage_root_id
  AND c.is_restricted <> (latest.status IN ('pending', 'cancel'))
RETURNING c.*;

-- name: ListRestrictedCreditsOlderThan :many
-- Credits that have been restricted for 30+ days and still fund a live
-- event — the cleanup cron's second deletion criterion (see
-- ListFlashEventsOlderThan in events.sql for the first).
SELECT * FROM credits
WHERE is_restricted = true AND restricted_since < $1 AND event_id IS NOT NULL;

-- name: ListCreditsForOrganization :many
-- The org's whole credit ledger — PlansService.ListBilling derives the
-- available-by-type summary from this rather than a separate aggregate
-- query, since the frontend also wants to show consumed/restricted rows.
SELECT * FROM credits WHERE organization_id = $1 ORDER BY created_at ASC;
