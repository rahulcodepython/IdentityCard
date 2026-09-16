package eventform

import (
    "context"
    "time"

    "identitycard-server/internal/pkg/postgres"
)

func (r *App) GetEventFormRepository(ctx context.Context, eventID string) (*EventFormDetails, error) {
    return postgres.QueryJSON[EventFormDetails](ctx, r.DB, GetEventFormByEventIDQuery, eventID)
}

func (r *App) CreateFromTemplateRepository(ctx context.Context, eventID, templateID, name string, maxApplicants int, expiresAt time.Time) (string, error) {
    var id string
    err := r.DB.QueryRow(ctx, CreateEventFormFromTemplateQuery, eventID, templateID, name, maxApplicants, expiresAt).Scan(&id)
    if err != nil {
        return "", postgres.MapPgError(err)
    }
    return id, nil
}

func (r *App) CreateFromScratchRepository(ctx context.Context, eventID, name string, fieldsJSON []byte, maxApplicants int, expiresAt time.Time) (string, error) {
    var id string
    err := r.DB.QueryRow(ctx, CreateEventFormFromScratchQuery, eventID, name, fieldsJSON, maxApplicants, expiresAt).Scan(&id)
    if err != nil {
        return "", postgres.MapPgError(err)
    }
    return id, nil
}

func (r *App) UpdateEventFormRepository(ctx context.Context, eventID string, name *string, fieldsJSON []byte, maxApplicants *int, expiresAt *time.Time) error {
    var fieldsArg any
    if fieldsJSON != nil {
        fieldsArg = string(fieldsJSON)
    }
    var id string
    err := r.DB.QueryRow(ctx, UpdateEventFormQuery, eventID, name, fieldsArg, maxApplicants, expiresAt).Scan(&id)
    if err != nil {
        return postgres.MapPgError(err)
    }
    return nil
}

func (r *App) LockEventFormRepository(ctx context.Context, eventID string) error {
    var id string
    err := r.DB.QueryRow(ctx, LockEventFormQuery, eventID).Scan(&id)
    if err != nil {
        return postgres.MapPgError(err)
    }
    return nil
}

func (r *App) DeleteEventFormRepository(ctx context.Context, eventID string) error {
    var id string
    err := r.DB.QueryRow(ctx, DeleteEventFormQuery, eventID).Scan(&id)
    if err != nil {
        return postgres.MapPgError(err)
    }
    return nil
}

func (r *App) CountApplicantsRepository(ctx context.Context, eventID string) (int, error) {
    var count int
    err := r.DB.QueryRow(ctx, CountApplicantsForEventQuery, eventID).Scan(&count)
    if err != nil {
        return 0, postgres.MapPgError(err)
    }
    return count, nil
}

func (r *App) CheckLockStatusRepository(ctx context.Context, eventID string) (bool, error) {
    var isLocked bool
    err := r.DB.QueryRow(ctx, CheckEventFormLockStatusQuery, eventID).Scan(&isLocked)
    if err != nil {
        return false, postgres.MapPgError(err)
    }
    return isLocked, nil
}
