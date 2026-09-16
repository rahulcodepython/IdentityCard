package dates

import (
    "context"
    "strings"
    "time"
)

// ListByMonthService validates parameters and fetches event dates, or all dates if no month is provided.
func (s *App) ListByMonthService(ctx context.Context, eventID string, month string) ([]EventDate, error) {
    targetMonth := strings.TrimSpace(month)
    if targetMonth != "" {
        if _, err := time.Parse("2006-01", targetMonth); err != nil {
            return nil, ErrInvalidDateFormat
        }
    }

    return s.ListByMonthRepository(ctx, eventID, targetMonth)
}

// BulkUpsertService updates event dates.
func (s *App) BulkUpsertService(ctx context.Context, eventID string, req BulkUpsertEventDatesRequest) ([]EventDate, error) {
    return s.BulkUpsertRepository(ctx, eventID, req.Dates)
}

// OverwrideService replaces all event dates with the provided batch.
func (s *App) OverwrideService(ctx context.Context, eventID string, req BulkUpsertEventDatesRequest) ([]EventDate, error) {
    return s.OverwrideRepository(ctx, eventID, req.Dates)
}

// BulkDeleteService deletes specified dates for an event.
func (s *App) BulkDeleteService(ctx context.Context, eventID string, req BulkDeleteEventDatesRequest) error {
    return s.BulkDeleteRepository(ctx, eventID, req.Dates)
}
