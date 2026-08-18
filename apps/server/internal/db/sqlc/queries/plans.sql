-- name: ListPlans :many
-- Ordered explicitly by tier rather than created_at: rows seeded in the
-- same migration share one transaction timestamp, so created_at can't be
-- relied on to break ties in insertion order.
SELECT * FROM plans ORDER BY
    CASE kind WHEN 'flash' THEN 0 ELSE 1 END,
    CASE tier
        WHEN 'flash' THEN 0
        WHEN 'very_small' THEN 1
        WHEN 'small' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'large' THEN 4
        WHEN 'yearly' THEN 5
        WHEN 'recurring' THEN 6
        WHEN 'custom' THEN 7
        ELSE 99
    END;

-- name: GetPlanByCode :one
SELECT * FROM plans WHERE code = $1;
