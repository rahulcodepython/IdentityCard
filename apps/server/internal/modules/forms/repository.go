// Package forms owns shareable public sign-up links (event_forms): admin
// CRUD on them, plus the unauthenticated read/submit path the link itself
// serves. Person creation is delegated to people.Service — this package
// only owns capacity accounting.
package forms

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{q: r.q.WithTx(tx)}
}

func (r *Repository) Create(ctx context.Context, orgID, eventID uuid.UUID, subEventID pgtype.UUID, token string, capacity pgtype.Int4) (dbgen.EventForm, error) {
	return r.q.CreateEventForm(ctx, dbgen.CreateEventFormParams{
		OrganizationID: orgID, EventID: eventID, SubEventID: subEventID, Token: token, Capacity: capacity,
	})
}

func (r *Repository) List(ctx context.Context, orgID, eventID uuid.UUID) ([]dbgen.EventForm, error) {
	return r.q.ListEventForms(ctx, dbgen.ListEventFormsParams{EventID: eventID, OrganizationID: orgID})
}

func (r *Repository) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (dbgen.EventForm, error) {
	return r.q.GetEventForm(ctx, dbgen.GetEventFormParams{ID: id, EventID: eventID, OrganizationID: orgID})
}

func (r *Repository) Update(ctx context.Context, orgID, eventID, id uuid.UUID, capacity pgtype.Int4, isActive bool) (dbgen.EventForm, error) {
	return r.q.UpdateEventForm(ctx, dbgen.UpdateEventFormParams{
		ID: id, EventID: eventID, OrganizationID: orgID, Capacity: capacity, IsActive: isActive,
	})
}

// Delete reports whether a row was actually deleted.
func (r *Repository) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
	n, err := r.q.DeleteEventForm(ctx, dbgen.DeleteEventFormParams{ID: id, EventID: eventID, OrganizationID: orgID})
	return n > 0, err
}

func (r *Repository) GetPublicByToken(ctx context.Context, token string) (dbgen.GetPublicFormByTokenRow, error) {
	return r.q.GetPublicFormByToken(ctx, token)
}

// IncrementSubmissions atomically charges one capacity slot; a returned
// pgx.ErrNoRows means the form is full or inactive.
func (r *Repository) IncrementSubmissions(ctx context.Context, id uuid.UUID) (dbgen.EventForm, error) {
	return r.q.IncrementEventFormSubmissions(ctx, id)
}

func (r *Repository) DecrementSubmissions(ctx context.Context, id uuid.UUID) error {
	return r.q.DecrementEventFormSubmissions(ctx, id)
}
