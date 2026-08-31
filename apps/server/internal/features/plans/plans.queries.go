package plans

const (
	ListPlansQuery = `
        SELECT COALESCE(jsonb_agg(row_to_json(p)), '[]'::jsonb)
        FROM (
            SELECT * FROM plans
            ORDER BY
                CASE kind WHEN 'flash' THEN 0 WHEN 'base' THEN 1 WHEN 'custom' THEN 2 WHEN 'unlimited' THEN 3 ELSE 99 END,
                CASE billing_cycle WHEN 'one_time' THEN 0 WHEN 'monthly' THEN 1 WHEN 'yearly' THEN 2 ELSE 99 END
        ) p;
    `

	GetPlanByCodeQuery = `
        SELECT row_to_json(p) FROM plans p WHERE p.code = $1;
    `

	GetPlanByIDQuery = `
        SELECT row_to_json(p) FROM plans p WHERE p.id = $1;
    `

	CreateBillingLineageRootQuery = `
        WITH new_billing AS (
            INSERT INTO billing (organization_id, plan_id, billing_number, period_start, period_end, status, amount)
            VALUES ($1, $2, 1, $3, $4, $5, $6)
            RETURNING id, organization_id, plan_id, billing_number, period_start, period_end, status, amount, paid_at, transaction_id, created_at
        ), updated AS (
            UPDATE billing SET lineage_root_id = new_billing.id
            FROM new_billing
            WHERE billing.id = new_billing.id
            RETURNING billing.*
        )
        SELECT row_to_json(u) FROM updated u;
    `

	CreateBillingRenewalQuery = `
        WITH ins AS (
            INSERT INTO billing (organization_id, plan_id, lineage_root_id, billing_number, period_start, period_end, status, amount)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `

	GetLatestBillingForLineageQuery = `
        SELECT row_to_json(b) FROM (
            SELECT * FROM billing WHERE lineage_root_id = $1 ORDER BY period_start DESC LIMIT 1
        ) b;
    `

	ListLatestBillingForOrganizationQuery = `
        SELECT COALESCE(jsonb_agg(row_to_json(b)), '[]'::jsonb) FROM (
            SELECT DISTINCT ON (lineage_root_id) *
            FROM billing
            WHERE organization_id = $1
            ORDER BY lineage_root_id, period_start DESC
        ) b;
    `

	MarkBillingPaidQuery = `
        WITH upd AS (
            UPDATE billing SET status = 'paid', paid_at = now(), transaction_id = $2
            WHERE id = $1
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	CancelBillingLineageQuery = `
        WITH upd AS (
            UPDATE billing SET status = 'cancel'
            WHERE id = (SELECT b.id FROM billing b WHERE b.lineage_root_id = $1 ORDER BY b.period_start DESC LIMIT 1)
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	UpgradeBillingPlanQuery = `
        WITH upd AS (
            UPDATE billing SET plan_id = $2, amount = $3
            WHERE id = (SELECT b.id FROM billing b WHERE b.lineage_root_id = $1 ORDER BY b.period_start DESC LIMIT 1)
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	FlipLapsedBillingToPendingQuery = `
        WITH upd AS (
            UPDATE billing b
            SET status = 'pending'
            WHERE b.status = 'active' AND b.period_end < $1::date
              AND b.id = (SELECT id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
            RETURNING *
        )
        SELECT COALESCE(jsonb_agg(row_to_json(upd)), '[]'::jsonb) FROM upd;
    `

	CreateCreditQuery = `
        WITH ins AS (
            INSERT INTO credits (organization_id, billing_id, type)
            VALUES ($1, $2, $3)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `

	FindAvailableCreditQuery = `
        SELECT row_to_json(c) FROM (
            SELECT * FROM credits
            WHERE organization_id = $1
              AND type = ANY($2::text[])
              AND event_id IS NULL
              AND is_restricted = false
              AND (type <> 'flash' OR created_at::date = CURRENT_DATE)
            ORDER BY created_at ASC
            LIMIT 1
            FOR UPDATE
        ) c;
    `

	ConsumeCreditQuery = `
        WITH upd AS (
            UPDATE credits SET event_id = $2 WHERE id = $1 RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	GetCreditForEventQuery = `
        SELECT row_to_json(c) FROM (
            SELECT * FROM credits WHERE event_id = $1
        ) c;
    `

	GetActiveUnlimitedLineageRootQuery = `
        SELECT b.lineage_root_id FROM billing b
        JOIN plans p ON p.id = b.plan_id
        WHERE b.organization_id = $1 AND p.kind = 'unlimited'
          AND b.status IN ('active', 'paid')
          AND b.id = (SELECT id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
        LIMIT 1;
    `

	ListCreditsForOrganizationQuery = `
        SELECT COALESCE(jsonb_agg(row_to_json(c)), '[]'::jsonb) FROM (
            SELECT * FROM credits WHERE organization_id = $1 ORDER BY created_at ASC
        ) c;
    `

	SyncCreditRestrictionQuery = `
        WITH upd AS (
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
            RETURNING c.*
        )
        SELECT COALESCE(jsonb_agg(row_to_json(upd)), '[]'::jsonb) FROM upd;
    `

	ListRestrictedCreditsOlderThanQuery = `
        SELECT COALESCE(jsonb_agg(row_to_json(c)), '[]'::jsonb) FROM (
            SELECT * FROM credits
            WHERE is_restricted = true AND restricted_since < $1 AND event_id IS NOT NULL
        ) c;
    `

	CreateTransactionQuery = `
        WITH ins AS (
            INSERT INTO transactions (organization_id, plan_id, amount, currency)
            VALUES ($1, $2, $3, $4)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `
)
