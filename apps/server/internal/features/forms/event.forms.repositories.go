package forms

import (
    "context"
    "time"

    "identitycard-server/internal/pkg/postgres"
)

// UpdateEventFormQueryResult carries the outcome of atomic form update CTE.
type UpdateEventFormQueryResult struct {
    Exists   bool              `json:"exists"`
    IsLocked bool              `json:"is_locked"`
    Form     *EventFormDetails `json:"form"`
}

// LockEventFormQueryResult carries the outcome of atomic form lock CTE.
type LockEventFormQueryResult struct {
    Exists    bool              `json:"exists"`
    WasLocked bool              `json:"was_locked"`
    IsExpired bool              `json:"is_expired"`
    Form      *EventFormDetails `json:"form"`
}

// DeleteEventFormQueryResult carries the outcome of atomic form deletion CTE.
type DeleteEventFormQueryResult struct {
    Exists  bool `json:"exists"`
    Deleted bool `json:"deleted"`
}

func (r *App) GetEventFormRepository(ctx context.Context, eventID string) (*EventFormDetails, error) {
    return postgres.QueryJSON[EventFormDetails](ctx, r.DB, GetEventFormByEventIDQuery, eventID)
}

func (r *App) CreateFromTemplateRepository(ctx context.Context, eventID, templateID, name string, maxApplicants int, expiresAt time.Time) (*EventFormDetails, error) {
    return postgres.QueryJSON[EventFormDetails](ctx, r.DB, CreateEventFormFromTemplateQuery, eventID, templateID, name, maxApplicants, expiresAt)
}

func (r *App) CreateFromScratchRepository(ctx context.Context, eventID, name string, fieldsJSON []byte, maxApplicants int, expiresAt time.Time) (*EventFormDetails, error) {
    return postgres.QueryJSON[EventFormDetails](ctx, r.DB, CreateEventFormFromScratchQuery, eventID, name, fieldsJSON, maxApplicants, expiresAt)
}

func (r *App) UpdateEventFormRepository(ctx context.Context, eventID string, name *string, fieldsJSON []byte, maxApplicants *int, expiresAt *time.Time) (*UpdateEventFormQueryResult, error) {
    var fieldsArg any
    if fieldsJSON != nil {
        fieldsArg = string(fieldsJSON)
    }
    return postgres.QueryJSON[UpdateEventFormQueryResult](ctx, r.DB, UpdateEventFormQuery, eventID, name, fieldsArg, maxApplicants, expiresAt)
}

func (r *App) LockEventFormRepository(ctx context.Context, eventID string) (*LockEventFormQueryResult, error) {
    return postgres.QueryJSON[LockEventFormQueryResult](ctx, r.DB, LockEventFormQuery, eventID)
}

func (r *App) DeleteEventFormRepository(ctx context.Context, eventID string) (*DeleteEventFormQueryResult, error) {
    return postgres.QueryJSON[DeleteEventFormQueryResult](ctx, r.DB, DeleteEventFormQuery, eventID)
}
