package jobs

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/pkg/mailer"
    "identitycard-server/internal/pkg/storage"
)

type milestoneNotification struct {
    LineageRootID    uuid.UUID
    OrganizationID   uuid.UUID
    OrgName          string
    Email            string
    UserName         string
    PeriodEnd        time.Time
    PruneDate        time.Time
    PlanName         string
    NotificationType string
}

// NewDailyMaintenanceTask constructs a periodic task for the jobs scheduler.
func NewDailyMaintenanceTask(pool *pgxpool.Pool, store *storage.Storage, mail *mailer.Mailer) Task {
    return func(ctx context.Context) error {
        return RunDailyMaintenance(ctx, pool, store, mail)
    }
}

// RunDailyMaintenance coordinates the daily lifecycle tasks:
// 1. Send consolidated 7-bucket lifecycle notification emails via Resend Batch.
// 2. Hard prune expired organizations (S3 assets and cascaded database entities).
// 3. Synchronize billing statuses and credit restrictions.
func RunDailyMaintenance(ctx context.Context, pool *pgxpool.Pool, store *storage.Storage, mail *mailer.Mailer) error {
    slog.Info("starting daily maintenance sweep")

    // 1. Notification Sweep
    if err := runNotificationSweep(ctx, pool, mail); err != nil {
        slog.Error("maintenance: notification sweep error", "error", err)
    }

    // 2. Hard Pruning Sweep
    if err := runHardPruningSweep(ctx, pool, store); err != nil {
        slog.Error("maintenance: hard pruning sweep error", "error", err)
    }

    // 3. Synchronize active billing periods that have lapsed
    if err := syncLapsedBilling(ctx, pool); err != nil {
        slog.Error("maintenance: sync lapsed billing error", "error", err)
    }

    slog.Info("completed daily maintenance sweep")
    return nil
}

func runNotificationSweep(ctx context.Context, pool *pgxpool.Pool, mail *mailer.Mailer) error {
    const milestoneQuery = `
        WITH milestone_candidates AS (
            -- 15 days before expiry
            SELECT b.lineage_root_id, b.organization_id, o.name AS organization_name, u.email, u.name AS user_name,
                   b.period_end, b.prune_date, p.name AS plan_name, 'expiry_15d' AS notification_type
            FROM billing b
            JOIN organizations o ON o.id = b.organization_id
            JOIN organization_members om ON om.organization_id = o.id AND 'super_admin' = ANY(om.roles)
            JOIN users u ON u.id = om.user_id
            JOIN plans p ON p.id = b.plan_id
            WHERE b.id = (SELECT b2.id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
              AND b.status IN ('active', 'paid')
              AND b.period_end = CURRENT_DATE + INTERVAL '15 days'

            UNION ALL

            -- 5 days before expiry
            SELECT b.lineage_root_id, b.organization_id, o.name AS organization_name, u.email, u.name AS user_name,
                   b.period_end, b.prune_date, p.name AS plan_name, 'expiry_5d' AS notification_type
            FROM billing b
            JOIN organizations o ON o.id = b.organization_id
            JOIN organization_members om ON om.organization_id = o.id AND 'super_admin' = ANY(om.roles)
            JOIN users u ON u.id = om.user_id
            JOIN plans p ON p.id = b.plan_id
            WHERE b.id = (SELECT b2.id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
              AND b.status IN ('active', 'paid')
              AND b.period_end = CURRENT_DATE + INTERVAL '5 days'

            UNION ALL

            -- 1 day before expiry
            SELECT b.lineage_root_id, b.organization_id, o.name AS organization_name, u.email, u.name AS user_name,
                   b.period_end, b.prune_date, p.name AS plan_name, 'expiry_1d' AS notification_type
            FROM billing b
            JOIN organizations o ON o.id = b.organization_id
            JOIN organization_members om ON om.organization_id = o.id AND 'super_admin' = ANY(om.roles)
            JOIN users u ON u.id = om.user_id
            JOIN plans p ON p.id = b.plan_id
            WHERE b.id = (SELECT b2.id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
              AND b.status IN ('active', 'paid')
              AND b.period_end = CURRENT_DATE + INTERVAL '1 day'

            UNION ALL

            -- Expiry today
            SELECT b.lineage_root_id, b.organization_id, o.name AS organization_name, u.email, u.name AS user_name,
                   b.period_end, b.prune_date, p.name AS plan_name, 'expired_today' AS notification_type
            FROM billing b
            JOIN organizations o ON o.id = b.organization_id
            JOIN organization_members om ON om.organization_id = o.id AND 'super_admin' = ANY(om.roles)
            JOIN users u ON u.id = om.user_id
            JOIN plans p ON p.id = b.plan_id
            WHERE b.id = (SELECT b2.id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
              AND b.status IN ('active', 'paid', 'pending')
              AND b.period_end = CURRENT_DATE

            UNION ALL

            -- 15 days before prune
            SELECT b.lineage_root_id, b.organization_id, o.name AS organization_name, u.email, u.name AS user_name,
                   b.period_end, b.prune_date, p.name AS plan_name, 'prune_15d' AS notification_type
            FROM billing b
            JOIN organizations o ON o.id = b.organization_id
            JOIN organization_members om ON om.organization_id = o.id AND 'super_admin' = ANY(om.roles)
            JOIN users u ON u.id = om.user_id
            JOIN plans p ON p.id = b.plan_id
            WHERE b.id = (SELECT b2.id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
              AND b.status IN ('pending', 'cancel')
              AND b.prune_date = CURRENT_DATE + INTERVAL '15 days'

            UNION ALL

            -- 5 days before prune
            SELECT b.lineage_root_id, b.organization_id, o.name AS organization_name, u.email, u.name AS user_name,
                   b.period_end, b.prune_date, p.name AS plan_name, 'prune_5d' AS notification_type
            FROM billing b
            JOIN organizations o ON o.id = b.organization_id
            JOIN organization_members om ON om.organization_id = o.id AND 'super_admin' = ANY(om.roles)
            JOIN users u ON u.id = om.user_id
            JOIN plans p ON p.id = b.plan_id
            WHERE b.id = (SELECT b2.id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
              AND b.status IN ('pending', 'cancel')
              AND b.prune_date = CURRENT_DATE + INTERVAL '5 days'

            UNION ALL

            -- 1 day before prune
            SELECT b.lineage_root_id, b.organization_id, o.name AS organization_name, u.email, u.name AS user_name,
                   b.period_end, b.prune_date, p.name AS plan_name, 'prune_1d' AS notification_type
            FROM billing b
            JOIN organizations o ON o.id = b.organization_id
            JOIN organization_members om ON om.organization_id = o.id AND 'super_admin' = ANY(om.roles)
            JOIN users u ON u.id = om.user_id
            JOIN plans p ON p.id = b.plan_id
            WHERE b.id = (SELECT b2.id FROM billing b2 WHERE b2.lineage_root_id = b.lineage_root_id ORDER BY b2.period_start DESC LIMIT 1)
              AND b.status IN ('pending', 'cancel')
              AND b.prune_date = CURRENT_DATE + INTERVAL '1 day'
        )
        SELECT m.lineage_root_id, m.organization_id, m.organization_name, m.email, m.user_name,
               m.period_end, m.prune_date, m.plan_name, m.notification_type
        FROM milestone_candidates m
        LEFT JOIN billing_lifecycle_notifications bln 
          ON bln.lineage_root_id = m.lineage_root_id AND bln.notification_type = m.notification_type
        WHERE bln.id IS NULL;
    `

    rows, err := pool.Query(ctx, milestoneQuery)
    if err != nil {
        return fmt.Errorf("failed to query lifecycle milestones: %w", err)
    }
    defer rows.Close()

    var notifications []milestoneNotification
    for rows.Next() {
        var n milestoneNotification
        if err := rows.Scan(
            &n.LineageRootID,
            &n.OrganizationID,
            &n.OrgName,
            &n.Email,
            &n.UserName,
            &n.PeriodEnd,
            &n.PruneDate,
            &n.PlanName,
            &n.NotificationType,
        ); err != nil {
            return fmt.Errorf("failed to scan milestone notification: %w", err)
        }
        notifications = append(notifications, n)
    }

    if len(notifications) == 0 {
        slog.Info("maintenance: no lifecycle notifications to send today")
        return nil
    }

    var batch []mailer.BatchMessage
    for _, n := range notifications {
        subject, body := formatLifecycleEmail(n)
        batch = append(batch, mailer.BatchMessage{
            To:      n.Email,
            Subject: subject,
            Body:    body,
        })
    }

    if mail != nil && len(batch) > 0 {
        if err := mail.SendBatch(ctx, batch); err != nil {
            return fmt.Errorf("failed to dispatch notification batch: %w", err)
        }
    }

    // Record delivery idempotently
    const recordQuery = `
        INSERT INTO billing_lifecycle_notifications (organization_id, lineage_root_id, notification_type)
        VALUES ($1, $2, $3)
        ON CONFLICT (lineage_root_id, notification_type) DO NOTHING;
    `
    for _, n := range notifications {
        if _, err := pool.Exec(ctx, recordQuery, n.OrganizationID, n.LineageRootID, n.NotificationType); err != nil {
            slog.Error("maintenance: failed to log notification record", "lineage_id", n.LineageRootID, "type", n.NotificationType, "error", err)
        }
    }

    slog.Info("maintenance: dispatched lifecycle notifications", "count", len(notifications))
    return nil
}

func runHardPruningSweep(ctx context.Context, pool *pgxpool.Pool, store *storage.Storage) error {
    const expiredOrgsQuery = `
        SELECT DISTINCT b.organization_id
        FROM billing b
        WHERE b.prune_date <= CURRENT_DATE
          AND b.status IN ('pending', 'cancel')
          AND NOT EXISTS (
              SELECT 1 FROM billing b2
              WHERE b2.organization_id = b.organization_id
                AND b2.status IN ('active', 'paid')
                AND b2.period_end >= CURRENT_DATE
          );
    `

    rows, err := pool.Query(ctx, expiredOrgsQuery)
    if err != nil {
        return fmt.Errorf("failed to query expired orgs for pruning: %w", err)
    }
    defer rows.Close()

    var orgIDs []uuid.UUID
    for rows.Next() {
        var id uuid.UUID
        if err := rows.Scan(&id); err != nil {
            return fmt.Errorf("failed to scan expired org id: %w", err)
        }
        orgIDs = append(orgIDs, id)
    }

    if len(orgIDs) == 0 {
        slog.Info("maintenance: no organizations eligible for pruning")
        return nil
    }

    // 1. Purge S3 assets for events in these organizations
    if store != nil {
        const assetsQuery = `
            SELECT image_object_key, organizer_signature_object_key
            FROM events
            WHERE organization_id = ANY($1)
              AND (image_object_key IS NOT NULL OR organizer_signature_object_key IS NOT NULL);
        `
        assetRows, err := pool.Query(ctx, assetsQuery, orgIDs)
        if err == nil {
            var keys []string
            for assetRows.Next() {
                var imgKey, sigKey *string
                if err := assetRows.Scan(&imgKey, &sigKey); err == nil {
                    if imgKey != nil && *imgKey != "" {
                        keys = append(keys, *imgKey)
                    }
                    if sigKey != nil && *sigKey != "" {
                        keys = append(keys, *sigKey)
                    }
                }
            }
            assetRows.Close()

            for _, key := range keys {
                if err := store.Delete(ctx, key); err != nil {
                    slog.Warn("maintenance: failed to delete S3 asset", "key", key, "error", err)
                }
            }
        }
    }

    // 2. Cascade delete events belonging to expired organizations
    const deleteEventsQuery = `
        DELETE FROM events
        WHERE organization_id = ANY($1);
    `
    tag, err := pool.Exec(ctx, deleteEventsQuery, orgIDs)
    if err != nil {
        return fmt.Errorf("failed to delete events for pruned orgs: %w", err)
    }

    // 3. Restrict any lingering credits
    const restrictCreditsQuery = `
        UPDATE credits
        SET is_restricted = true,
            restricted_since = now()
        WHERE organization_id = ANY($1) AND NOT is_restricted;
    `
    _, _ = pool.Exec(ctx, restrictCreditsQuery, orgIDs)

    slog.Info("maintenance: hard pruning completed", "pruned_orgs", len(orgIDs), "deleted_events", tag.RowsAffected())
    return nil
}

func syncLapsedBilling(ctx context.Context, pool *pgxpool.Pool) error {
    const syncQuery = `
        UPDATE billing
        SET status = 'pending'
        WHERE status = 'active' AND period_end < CURRENT_DATE;
    `
    tag, err := pool.Exec(ctx, syncQuery)
    if err != nil {
        return fmt.Errorf("failed to sync lapsed billing: %w", err)
    }
    if tag.RowsAffected() > 0 {
        slog.Info("maintenance: flipped lapsed billing to pending", "rows", tag.RowsAffected())
    }
    return nil
}

func formatLifecycleEmail(n milestoneNotification) (string, string) {
    switch n.NotificationType {
    case "expiry_15d":
        subject := fmt.Sprintf("Subscription Expiry in 15 Days - %s", n.OrgName)
        body := fmt.Sprintf("Hi %s,\n\nYour subscription for %s (%s plan) will expire on %s (in 15 days). Please renew your subscription to maintain uninterrupted access.\n",
            n.UserName, n.OrgName, n.PlanName, n.PeriodEnd.Format("2006-01-02"))
        return subject, body

    case "expiry_5d":
        subject := fmt.Sprintf("Urgent: Subscription Expiry in 5 Days - %s", n.OrgName)
        body := fmt.Sprintf("Hi %s,\n\nYour subscription for %s (%s plan) will expire on %s (in 5 days). Please renew now to avoid entering the grace period.\n",
            n.UserName, n.OrgName, n.PlanName, n.PeriodEnd.Format("2006-01-02"))
        return subject, body

    case "expiry_1d":
        subject := fmt.Sprintf("Action Required: Subscription Expires Tomorrow - %s", n.OrgName)
        body := fmt.Sprintf("Hi %s,\n\nYour subscription for %s (%s plan) expires tomorrow (%s). Renew today to avoid disruption.\n",
            n.UserName, n.OrgName, n.PlanName, n.PeriodEnd.Format("2006-01-02"))
        return subject, body

    case "expired_today":
        subject := fmt.Sprintf("Subscription Expired - Grace Period Started - %s", n.OrgName)
        body := fmt.Sprintf("Hi %s,\n\nYour subscription for %s (%s plan) has expired today. Your organization has entered a 30-day grace period ending on %s. During this time, event creation and card generation are paused until you renew.\n",
            n.UserName, n.OrgName, n.PlanName, n.PruneDate.Format("2006-01-02"))
        return subject, body

    case "prune_15d":
        subject := fmt.Sprintf("Warning: Permanent Data Deletion in 15 Days - %s", n.OrgName)
        body := fmt.Sprintf("Hi %s,\n\nYour grace period for %s will end in 15 days on %s. If you do not renew before this date, all event records, attendee cards, and media assets will be permanently deleted.\n",
            n.UserName, n.OrgName, n.PruneDate.Format("2006-01-02"))
        return subject, body

    case "prune_5d":
        subject := fmt.Sprintf("URGENT: Permanent Data Deletion in 5 Days - %s", n.OrgName)
        body := fmt.Sprintf("Hi %s,\n\nThis is an urgent reminder that all data for %s will be permanently pruned and deleted in 5 days on %s. Renew immediately to preserve your data.\n",
            n.UserName, n.OrgName, n.PruneDate.Format("2006-01-02"))
        return subject, body

    case "prune_1d":
        subject := fmt.Sprintf("FINAL NOTICE: Permanent Data Deletion Tomorrow - %s", n.OrgName)
        body := fmt.Sprintf("Hi %s,\n\nThis is your final notice. Tomorrow (%s), all event data, attendee records, and uploaded files for %s will be irrevocably deleted. Renew now if you wish to retain your data.\n",
            n.UserName, n.PruneDate.Format("2006-01-02"), n.OrgName)
        return subject, body

    default:
        return fmt.Sprintf("Subscription Update - %s", n.OrgName), fmt.Sprintf("Hello %s, please check your subscription status for %s.", n.UserName, n.OrgName)
    }
}
