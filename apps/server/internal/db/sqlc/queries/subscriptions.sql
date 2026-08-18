-- name: CreateSubscription :one
INSERT INTO subscriptions (organization_id, plan_id, status)
VALUES ($1, $2, 'active')
RETURNING *;

-- name: GetActiveSubscriptionForOrganization :one
SELECT * FROM subscriptions
WHERE organization_id = $1 AND status = 'active'
ORDER BY started_at DESC
LIMIT 1;
