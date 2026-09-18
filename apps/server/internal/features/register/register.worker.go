package register

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "strings"
    "time"

    "github.com/redis/go-redis/v9"
)

// StartWorker initiates the background Redis Stream consumer group worker.
func (s *App) StartWorker(ctx context.Context) {
    if s.RDB == nil {
        slog.Warn("redis client not provided, background submission worker not started")
        return
    }

    // 1. Create consumer group if not already existing
    err := s.RDB.XGroupCreateMkStream(ctx, StreamEventSubmissions, GroupRegisterWorkers, "$").Err()
    if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
        slog.Warn("failed to create redis consumer group", "error", err)
    }

    consumerName := fmt.Sprintf("consumer-%d", time.Now().UnixNano()%10000)

    go func() {
        slog.Info("starting background form submission worker", "consumer", consumerName)

        // Process pending unacknowledged messages first (crash recovery)
        s.drainStreamBatch(ctx, consumerName, "0")

        // Continuous ingestion loop
        for {
            select {
            case <-ctx.Done():
                slog.Info("stopping background form submission worker", "consumer", consumerName)
                return
            default:
                s.drainStreamBatch(ctx, consumerName, ">")
            }
        }
    }()
}

// drainStreamBatch reads up to 100 messages and writes them to PostgreSQL atomically.
func (s *App) drainStreamBatch(ctx context.Context, consumerName, streamPosition string) {
    streams, err := s.RDB.XReadGroup(ctx, &redis.XReadGroupArgs{
        Group:    GroupRegisterWorkers,
        Consumer: consumerName,
        Streams:  []string{StreamEventSubmissions, streamPosition},
        Count:    100,
        Block:    1000 * time.Millisecond,
    }).Result()

    if err != nil {
        if !errors.Is(err, redis.Nil) && !errors.Is(err, context.Canceled) {
            slog.Error("XReadGroup error", "error", err)
        }
        return
    }

    for _, stream := range streams {
        if len(stream.Messages) == 0 {
            continue
        }

        var submissions []QueuedSubmission
        var ackIDs []string

        for _, msg := range stream.Messages {
            payloadStr, ok := msg.Values["payload"].(string)
            if !ok {
                ackIDs = append(ackIDs, msg.ID)
                continue
            }

            var sub QueuedSubmission
            if err := json.Unmarshal([]byte(payloadStr), &sub); err != nil {
                slog.Error("failed to unmarshal queued submission", "error", err, "msg_id", msg.ID)
                ackIDs = append(ackIDs, msg.ID)
                continue
            }

            submissions = append(submissions, sub)
            ackIDs = append(ackIDs, msg.ID)
        }

        if len(submissions) > 0 {
            if err := s.persistBatchToDatabase(ctx, submissions); err != nil {
                slog.Error("failed to persist submission batch to database", "error", err, "count", len(submissions))
                // Do not ACK on failure so messages remain in Pending Entries List (PEL) to be retried
                time.Sleep(500 * time.Millisecond)
                continue
            }
        }

        if len(ackIDs) > 0 {
            if err := s.RDB.XAck(ctx, StreamEventSubmissions, GroupRegisterWorkers, ackIDs...).Err(); err != nil {
                slog.Error("failed to XACK processed messages", "error", err)
            }
        }
    }
}

// persistBatchToDatabase writes a batch of submissions into applicants and event_applicants tables.
func (s *App) persistBatchToDatabase(ctx context.Context, submissions []QueuedSubmission) error {
    tx, err := s.DB.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx) //nolint:errcheck

    for _, sub := range submissions {
        dataJSON, err := json.Marshal(sub.Data)
        if err != nil {
            dataJSON = []byte("{}")
        }

        // 1. Upsert applicant
        _, err = tx.Exec(ctx, `
            INSERT INTO applicants (id, name, email, phone, data, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5::jsonb, now(), now())
            ON CONFLICT (id) DO UPDATE 
            SET name = EXCLUDED.name, 
                phone = EXCLUDED.phone, 
                data = EXCLUDED.data, 
                updated_at = now()
        `, sub.UserID, sub.Name, sub.Email, sub.Phone, string(dataJSON))
        if err != nil {
            return fmt.Errorf("upsert applicant (%s): %w", sub.UserID, err)
        }

        // 2. Insert event_applicant with status = 'submitted'
        _, err = tx.Exec(ctx, `
            INSERT INTO event_applicants (event_id, user_id, email, created_at, status)
            VALUES ($1::uuid, $2, $3, now(), 'submitted')
            ON CONFLICT (event_id, email) DO NOTHING
        `, sub.EventID, sub.UserID, sub.Email)
        if err != nil {
            return fmt.Errorf("insert event_applicant (%s, %s): %w", sub.EventID, sub.Email, err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit batch tx: %w", err)
    }

    return nil
}
