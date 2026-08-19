-- name: ListPlans :many
SELECT * FROM plans ORDER BY
    CASE kind WHEN 'flash' THEN 0 WHEN 'base' THEN 1 WHEN 'custom' THEN 2 WHEN 'unlimited' THEN 3 ELSE 99 END,
    CASE billing_cycle WHEN 'one_time' THEN 0 WHEN 'monthly' THEN 1 WHEN 'yearly' THEN 2 ELSE 99 END;

-- name: GetPlanByCode :one
SELECT * FROM plans WHERE code = $1;

-- name: GetPlanByID :one
SELECT * FROM plans WHERE id = $1;
