package forms

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

func (a *App) CreateForm(ctx context.Context, orgID, eventID uuid.UUID, subEventID pgtype.UUID, token string, capacity pgtype.Int4) (dbgen.EventForm, error) {
	return a.queries.CreateEventForm(ctx, dbgen.CreateEventFormParams{
		OrganizationID: orgID, EventID: eventID, SubEventID: subEventID, Token: token, Capacity: capacity,
	})
}

func (a *App) ListForms(ctx context.Context, orgID, eventID uuid.UUID) ([]dbgen.EventForm, error) {
	return a.queries.ListEventForms(ctx, dbgen.ListEventFormsParams{EventID: eventID, OrganizationID: orgID})
}

func (a *App) GetForm(ctx context.Context, orgID, eventID, id uuid.UUID) (dbgen.EventForm, error) {
	return a.queries.GetEventForm(ctx, dbgen.GetEventFormParams{ID: id, EventID: eventID, OrganizationID: orgID})
}

func (a *App) UpdateForm(ctx context.Context, orgID, eventID, id uuid.UUID, capacity pgtype.Int4, isActive bool) (dbgen.EventForm, error) {
	return a.queries.UpdateEventForm(ctx, dbgen.UpdateEventFormParams{
		ID: id, EventID: eventID, OrganizationID: orgID, Capacity: capacity, IsActive: isActive,
	})
}

func (a *App) DeleteForm(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
	n, err := a.queries.DeleteEventForm(ctx, dbgen.DeleteEventFormParams{ID: id, EventID: eventID, OrganizationID: orgID})
	return n > 0, err
}

func (a *App) GetPublicByToken(ctx context.Context, token string) (dbgen.GetPublicFormByTokenRow, error) {
	return a.queries.GetPublicFormByToken(ctx, token)
}

func (a *App) IncrementSubmissions(ctx context.Context, id uuid.UUID) (dbgen.EventForm, error) {
	return a.queries.IncrementEventFormSubmissions(ctx, id)
}

func (a *App) DecrementSubmissions(ctx context.Context, id uuid.UUID) error {
	return a.queries.DecrementEventFormSubmissions(ctx, id)
}
