package jobs

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/pkg/mailer"
    "identitycard-server/internal/pkg/storage"
)

// NewDailyMaintenanceTask constructs a periodic task for the jobs scheduler.
func NewDailyMaintenanceTask(pool *pgxpool.Pool, store *storage.Storage, mail *mailer.Mailer) Task {
    return func(ctx context.Context) error {
        return RunDailyMaintenance(ctx, pool, store, mail)
    }
}

// RunDailyMaintenance coordinates the daily lifecycle tasks:
// 1. Synchronize annual subscriptions that have lapsed based on event count.
// 2. Clean up any orphaned media assets.
func RunDailyMaintenance(ctx context.Context, pool *pgxpool.Pool, store *storage.Storage, mail *mailer.Mailer) error {
    slog.Info("starting daily maintenance sweep")

    if err := syncLapsedAnnualBilling(ctx, pool); err != nil {
        slog.Error("maintenance: sync lapsed billing error", "error", err)
    }

    slog.Info("completed daily maintenance sweep")
    return nil
}

func syncLapsedAnnualBilling(ctx context.Context, pool *pgxpool.Pool) error {
    // 1. If an organization has events and period_end has passed, set annual_fee_status to 'past_due'
    const pastDueQuery = `
        UPDATE organization_billing
        SET annual_fee_status = 'past_due',
            updated_at = now()
        WHERE annual_fee_status = 'active'
          AND current_period_end < CURRENT_DATE
          AND (SELECT COUNT(*) FROM events WHERE organization_id = organization_billing.organization_id) > 0;
    `
    tag1, err := pool.Exec(ctx, pastDueQuery)
    if err != nil {
        return fmt.Errorf("failed to flip lapsed billing with events to past_due: %w", err)
    }
    if tag1.RowsAffected() > 0 {
        slog.Info("maintenance: set lapsed subscriptions to past_due", "count", tag1.RowsAffected())
    }

    // 2. If an organization has 0 events and period_end has passed, reset to 'free'
    const resetFreeQuery = `
        UPDATE organization_billing
        SET annual_fee_status = 'free',
            current_period_start = NULL,
            current_period_end = NULL,
            updated_at = now()
        WHERE annual_fee_status = 'active'
          AND current_period_end < CURRENT_DATE
          AND (SELECT COUNT(*) FROM events WHERE organization_id = organization_billing.organization_id) = 0;
    `
    tag2, err := pool.Exec(ctx, resetFreeQuery)
    if err != nil {
        return fmt.Errorf("failed to reset zero-event billing to free: %w", err)
    }
    if tag2.RowsAffected() > 0 {
        slog.Info("maintenance: reset zero-event lapsed subscriptions to free", "count", tag2.RowsAffected())
    }

    return nil
}
