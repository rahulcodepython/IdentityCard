package jobs

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgtype"

    "identitycard-server/internal/features/events"
    "identitycard-server/internal/features/plans"
)

// NewBillingTasks returns the two daily sweeps that enforce the credits/billing
// payment lifecycle:
//  1. syncBillingStatus: flip lapsed 'active' billing periods to 'pending',
//     then re-sync every credit's is_restricted from its lineage's current status.
//  2. cleanupExpiredEvents: hard-delete (a) flash events whose single day is
//     >30 days old, and (b) any event whose funding credit has been restricted
//     for >30 days (ON DELETE SET NULL in credits.event_id automatically frees
//     the credit once the event is gone — no extra bookkeeping needed).
func NewBillingTasks(plansApp *plans.App, eventsApp *events.App) []Task {
    return []Task{
        func(ctx context.Context) error { return syncBillingStatus(ctx, plansApp) },
        func(ctx context.Context) error { return cleanupExpiredEvents(ctx, plansApp, eventsApp) },
    }
}

func syncBillingStatus(ctx context.Context, plansApp *plans.App) error {
    return plansApp.SyncBillingStatus(ctx, time.Now())
}

func cleanupExpiredEvents(ctx context.Context, plansApp *plans.App, eventsApp *events.App) error {
    cutoff := time.Now().AddDate(0, 0, -30)
    cutoffDate := pgtype.Date{Time: cutoff, Valid: true}

    // (a) Flash events whose single day is older than 30 days.
    flashEvents, err := eventsApp.ListFlashOlderThan(ctx, cutoffDate)
    if err != nil {
        return err
    }
    for _, event := range flashEvents {
        if err := eventsApp.DeleteByID(ctx, event.ID); err != nil {
            return err
        }
    }

    // (b) Events whose funding credit has been restricted for >30 days.
    restrictedCredits, err := plansApp.ListRestrictedCreditsOlderThan(ctx, cutoff)
    if err != nil {
        return err
    }
    for _, credit := range restrictedCredits {
        if !credit.EventID.Valid {
            continue // already freed (ON DELETE SET NULL, or never linked)
        }
        eventID := credit.EventID.Bytes
        if err := eventsApp.DeleteByID(ctx, eventID); err != nil {
            return err
        }
        // credits.event_id is set to NULL by the FK ON DELETE SET NULL trigger
        // when the event is deleted — no extra update needed here.
    }
    return nil
}
