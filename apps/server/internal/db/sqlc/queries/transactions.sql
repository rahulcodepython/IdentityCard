-- name: CreateTransaction :one
INSERT INTO transactions (organization_id, plan_id, amount, currency)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListTransactionsForOrganization :many
SELECT * FROM transactions WHERE organization_id = $1 ORDER BY created_at DESC;
