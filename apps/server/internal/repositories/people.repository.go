// Package repositories: people owns attendee records for an event. All
// three ingestion paths — manual entry, CSV import, and public-form
// submission — funnel through the same upsert (see
// services.PeopleService.upsert): one person per (event_id, email,
// mobile), resubmission just updates their details.
package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type PeopleRepository struct {
	q *dbgen.Queries
}

func NewPeopleRepository(q *dbgen.Queries) *PeopleRepository {
	return &PeopleRepository{q: q}
}

func (r *PeopleRepository) WithTx(tx pgx.Tx) *PeopleRepository {
	return &PeopleRepository{q: r.q.WithTx(tx)}
}

// PeopleFields is the mutable, admin/CSV/public-submitted part of a person
// record — everything except identity (email/mobile, which key the
// upsert) and joined_at (which upsert semantics never overwrite).
type PeopleFields struct {
	Name     string
	ImageURL pgtype.Text
	Age      pgtype.Int2
	Gender   pgtype.Text
}

func (r *PeopleRepository) Upsert(ctx context.Context, orgID, eventID uuid.UUID, email, mobile string, f PeopleFields, joinedAt pgtype.Timestamptz) (dbgen.UpsertPersonRow, error) {
	return r.q.UpsertPerson(ctx, dbgen.UpsertPersonParams{
		OrganizationID: orgID,
		EventID:        eventID,
		Email:          email,
		Mobile:         mobile,
		Name:           f.Name,
		ImageUrl:       f.ImageURL,
		Age:            f.Age,
		Gender:         f.Gender,
		JoinedAt:       joinedAt,
	})
}

func (r *PeopleRepository) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (dbgen.Person, error) {
	return r.q.GetPerson(ctx, dbgen.GetPersonParams{ID: id, EventID: eventID, OrganizationID: orgID})
}

func (r *PeopleRepository) List(ctx context.Context, orgID, eventID uuid.UUID, subEventID pgtype.UUID, search pgtype.Text) ([]dbgen.Person, error) {
	return r.q.ListPeopleForEvent(ctx, dbgen.ListPeopleForEventParams{
		EventID: eventID, OrganizationID: orgID, SubEventID: subEventID, Search: search,
	})
}

func (r *PeopleRepository) ListByIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) ([]dbgen.Person, error) {
	return r.q.ListPeopleByIDs(ctx, dbgen.ListPeopleByIDsParams{EventID: eventID, OrganizationID: orgID, Ids: ids})
}

func (r *PeopleRepository) Update(ctx context.Context, orgID, eventID, id uuid.UUID, f PeopleFields) (dbgen.Person, error) {
	return r.q.UpdatePerson(ctx, dbgen.UpdatePersonParams{
		ID: id, EventID: eventID, OrganizationID: orgID,
		Name: f.Name, ImageUrl: f.ImageURL, Age: f.Age, Gender: f.Gender,
	})
}

// Delete reports whether a row was actually deleted.
func (r *PeopleRepository) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
	n, err := r.q.DeletePerson(ctx, dbgen.DeletePersonParams{ID: id, EventID: eventID, OrganizationID: orgID})
	return n > 0, err
}

func (r *PeopleRepository) ListSubEventIDs(ctx context.Context, personID uuid.UUID) ([]uuid.UUID, error) {
	return r.q.ListPersonSubEventIDs(ctx, personID)
}

func (r *PeopleRepository) AddToSubEvent(ctx context.Context, personID, subEventID uuid.UUID) error {
	return r.q.AddPersonToSubEvent(ctx, dbgen.AddPersonToSubEventParams{PersonID: personID, SubEventID: subEventID})
}

func (r *PeopleRepository) DeleteSubEventLinks(ctx context.Context, personID uuid.UUID) error {
	return r.q.DeletePersonSubEvents(ctx, personID)
}

func (r *PeopleRepository) MarkCardSent(ctx context.Context, personID uuid.UUID) error {
	return r.q.MarkCardSent(ctx, personID)
}
