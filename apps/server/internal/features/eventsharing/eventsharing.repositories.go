package eventsharing

import (
    "context"
    "time"

    "identitycard-server/internal/pkg/postgres"
)

func (r *App) GetSharingRepository(ctx context.Context, eventID string) (*EventSharingResponse, error) {
    return postgres.QueryJSON[EventSharingResponse](ctx, r.DB, GetEventSharingQuery, eventID)
}

func (r *App) UpsertSharingRepository(ctx context.Context, eventID, formID string, maxApplicants int, expiresAt time.Time, status EventFormStatus) (*UpsertSharingResult, error) {
    return postgres.QueryJSON[UpsertSharingResult](ctx, r.DB, UpsertSharingCTEQuery, eventID, formID, maxApplicants, expiresAt, status)
}
