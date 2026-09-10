package jobs

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/features/cards"
)

type OutboxWorker struct {
    pool     *pgxpool.Pool
    cardsApp *cards.App
    interval time.Duration
}

func NewOutboxWorker(pool *pgxpool.Pool, cardsApp *cards.App) *OutboxWorker {
    return &OutboxWorker{
        pool:     pool,
        cardsApp: cardsApp,
        interval: 2 * time.Second,
    }
}

type jobOutboxRecord struct {
    ID          uuid.UUID       `json:"id"`
    Queue       string          `json:"queue"`
    JobType     string          `json:"job_type"`
    Payload     json.RawMessage `json:"payload"`
    Attempts    int             `json:"attempts"`
    MaxAttempts int             `json:"max_attempts"`
}

type SendCardsPayload struct {
    OrgID   uuid.UUID `json:"org_id"`
    EventID uuid.UUID `json:"event_id"`
}

// Start polls and processes transactional outbox tasks until ctx is cancelled.
func (w *OutboxWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.processBatch(ctx)
        }
    }
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        job, err := w.claimNextJob(ctx)
        if err != nil {
            if err != pgx.ErrNoRows {
                slog.Error("outbox: failed to claim next job", "error", err)
            }
            return
        }

        w.executeJob(ctx, job)
    }
}

func (w *OutboxWorker) claimNextJob(ctx context.Context) (*jobOutboxRecord, error) {
    const claimQuery = `
        WITH next_job AS (
            SELECT id
            FROM job_outbox
            WHERE status IN ('pending', 'failed')
              AND run_at <= now()
              AND attempts < max_attempts
            ORDER BY run_at ASC
            LIMIT 1
            FOR UPDATE SKIP LOCKED
        )
        UPDATE job_outbox j
        SET status = 'processing',
            locked_at = now(),
            attempts = j.attempts + 1,
            updated_at = now()
        FROM next_job
        WHERE j.id = next_job.id
        RETURNING j.id, j.queue, j.job_type, j.payload, j.attempts, j.max_attempts;
    `

    var job jobOutboxRecord
    err := w.pool.QueryRow(ctx, claimQuery).Scan(
        &job.ID,
        &job.Queue,
        &job.JobType,
        &job.Payload,
        &job.Attempts,
        &job.MaxAttempts,
    )
    if err != nil {
        return nil, err
    }
    return &job, nil
}

func (w *OutboxWorker) executeJob(ctx context.Context, job *jobOutboxRecord) {
    var execErr error

    switch job.JobType {
    case "send_event_cards":
        var payload SendCardsPayload
        if err := json.Unmarshal(job.Payload, &payload); err != nil {
            execErr = fmt.Errorf("invalid send_event_cards payload: %w", err)
        } else if w.cardsApp == nil {
            execErr = fmt.Errorf("cards app not initialized in outbox worker")
        } else {
            execErr = w.cardsApp.SendForEvent(ctx, payload.OrgID, payload.EventID)
        }

    default:
        execErr = fmt.Errorf("unrecognized job_type: %s", job.JobType)
    }

    if execErr != nil {
        slog.Error("outbox: job failed", "job_id", job.ID, "type", job.JobType, "attempt", job.Attempts, "error", execErr)
        w.failJob(ctx, job.ID, job.Attempts, job.MaxAttempts, execErr)
        return
    }

    w.completeJob(ctx, job.ID)
}

func (w *OutboxWorker) completeJob(ctx context.Context, id uuid.UUID) {
    const completeQuery = `
        UPDATE job_outbox
        SET status = 'completed',
            updated_at = now()
        WHERE id = $1;
    `
    if _, err := w.pool.Exec(ctx, completeQuery, id); err != nil {
        slog.Error("outbox: failed to mark job completed", "job_id", id, "error", err)
    }
}

func (w *OutboxWorker) failJob(ctx context.Context, id uuid.UUID, attempts, maxAttempts int, jobErr error) {
    backoff := time.Duration(attempts*attempts) * 30 * time.Second
    nextRun := time.Now().Add(backoff)

    newStatus := "failed"
    if attempts >= maxAttempts {
        newStatus = "dead"
    }

    const failQuery = `
        UPDATE job_outbox
        SET status = $2,
            last_error = $3,
            run_at = $4,
            updated_at = now()
        WHERE id = $1;
    `
    if _, err := w.pool.Exec(ctx, failQuery, id, newStatus, jobErr.Error(), nextRun); err != nil {
        slog.Error("outbox: failed to record job failure", "job_id", id, "error", err)
    }
}
