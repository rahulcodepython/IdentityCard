package eventdates

import (
    "context"
    "encoding/json"

    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/pkg/postgres"
)

type bulkUpsertResult struct {
    Status string      `json:"status"`
    Data   []EventDate `json:"data"`
}

// ListByMonthRepository retrieves all event dates for an event within a specific month.
func (r *App) ListByMonthRepository(ctx context.Context, eventID string, month string) ([]EventDate, error) {
    return postgres.QueryJSONSlice[EventDate](ctx, r.DB, ListEventDatesByMonthQuery, eventID, month)
}

// BulkUpsertRepository validates and upserts dates within a single CTE network round-trip.
func (r *App) BulkUpsertRepository(ctx context.Context, eventID string, dates []EventDateItemInput) ([]EventDate, error) {
    datesJSON, err := json.Marshal(dates)
    if err != nil {
        return nil, err
    }

    result, err := postgres.QueryJSON[bulkUpsertResult](ctx, r.DB, BulkUpsertEventDatesQuery, eventID, string(datesJSON))
    if err != nil {
        return nil, err
    }
    if result == nil || result.Status == "not_found" {
        return nil, ErrEventNotFound
    }
    if result.Status == "date_out_of_range" {
        return nil, ErrDateOutOfRange
    }

    if result.Data == nil {
        return []EventDate{}, nil
    }
    return result.Data, nil
}

// OverwrideRepository removes all existing dates for an event and inserts the new dates in a single transaction.
func (r *App) OverwrideRepository(ctx context.Context, eventID string, dates []EventDateItemInput) ([]EventDate, error) {
    datesJSON, err := json.Marshal(dates)
    if err != nil {
        return nil, err
    }

    var inserted []EventDate
    err = postgres.WithTx(ctx, r.DB, func(tx pgx.Tx) error {
        var status string
        if err := tx.QueryRow(ctx, ValidateEventDatesQuery, eventID, string(datesJSON)).Scan(&status); err != nil {
            return postgres.MapPgError(err)
        }

        if status == "not_found" {
            return ErrEventNotFound
        }
        if status == "date_out_of_range" {
            return ErrDateOutOfRange
        }

        if _, err := tx.Exec(ctx, DeleteAllEventDatesQuery, eventID); err != nil {
            return postgres.MapPgError(err)
        }

        res, err := postgres.QueryJSONSlice[EventDate](ctx, tx, InsertEventDatesQuery, eventID, string(datesJSON))
        if err != nil {
            return err
        }
        inserted = res
        return nil
    })

    if err != nil {
        return nil, err
    }
    if inserted == nil {
        inserted = []EventDate{}
    }
    return inserted, nil
}

// BulkDeleteRepository removes the provided dates for an event in one network round-trip.
func (r *App) BulkDeleteRepository(ctx context.Context, eventID string, dates []string) error {
    _, err := r.DB.Exec(ctx, BulkDeleteEventDatesQuery, eventID, dates)
    if err != nil {
        return postgres.MapPgError(err)
    }
    return nil
}

