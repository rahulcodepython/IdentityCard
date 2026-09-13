package billing

const (
    GetOrganizationBillingQuery = `
        SELECT row_to_json(b) FROM (
            SELECT 
                organization_id,
                credit_balance,
                annual_fee_status,
                to_char(current_period_start, 'YYYY-MM-DD') AS current_period_start,
                to_char(current_period_end, 'YYYY-MM-DD') AS current_period_end,
                cancel_at_period_end,
                updated_at,
                created_at
            FROM organization_billing
            WHERE organization_id = $1
        ) b;
    `

    EnsureOrganizationBillingQuery = `
        INSERT INTO organization_billing (organization_id, credit_balance, annual_fee_status)
        VALUES ($1, $2, 'free')
        ON CONFLICT (organization_id) DO NOTHING;
    `

    CountEventsForOrgQuery = `
        SELECT COUNT(*)::int FROM events WHERE organization_id = $1;
    `

    ListBillingTransactionsQuery = `
        SELECT COALESCE(jsonb_agg(row_to_json(t)), '[]'::jsonb) FROM (
            SELECT 
                id,
                organization_id,
                type,
                credits_delta,
                amount,
                currency,
                event_id,
                description,
                created_at
            FROM billing_transactions
            WHERE organization_id = $1
            ORDER BY created_at DESC
            LIMIT 50
        ) t;
    `

    AddCreditsQuery = `
        UPDATE organization_billing
        SET credit_balance = credit_balance + $2,
            updated_at = now()
        WHERE organization_id = $1
        RETURNING credit_balance;
    `

    InsertBillingTransactionQuery = `
        INSERT INTO billing_transactions (organization_id, type, credits_delta, amount, currency, event_id, description)
        VALUES ($1, $2, $3, $4, $5, $6, $7);
    `

    RenewAnnualSubscriptionQuery = `
        UPDATE organization_billing
        SET annual_fee_status = 'active',
            current_period_start = CURRENT_DATE,
            current_period_end = CURRENT_DATE + INTERVAL '1 year',
            updated_at = now()
        WHERE organization_id = $1
        RETURNING *;
    `
)
