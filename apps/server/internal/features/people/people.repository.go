package people

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type PeopleFields struct {
	Name     string
	ImageURL pgtype.Text
	Age      pgtype.Int2
	Gender   pgtype.Text
}

func (a *App) UpsertPerson(ctx context.Context, tx pgx.Tx, orgID, eventID uuid.UUID, email, mobile string, f PeopleFields, joinedAt pgtype.Timestamptz) (dbgen.UpsertPersonRow, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.UpsertPerson(ctx, dbgen.UpsertPersonParams{
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

func (a *App) GetPerson(ctx context.Context, orgID, eventID, id uuid.UUID) (dbgen.Person, error) {
	return a.queries.GetPerson(ctx, dbgen.GetPersonParams{ID: id, EventID: eventID, OrganizationID: orgID})
}

func (a *App) ListPeople(ctx context.Context, orgID, eventID uuid.UUID, subEventID pgtype.UUID, search pgtype.Text) ([]dbgen.Person, error) {
	return a.queries.ListPeopleForEvent(ctx, dbgen.ListPeopleForEventParams{
		EventID: eventID, OrganizationID: orgID, SubEventID: subEventID, Search: search,
	})
}

func (a *App) ListPeopleByIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) ([]dbgen.Person, error) {
	return a.queries.ListPeopleByIDs(ctx, dbgen.ListPeopleByIDsParams{EventID: eventID, OrganizationID: orgID, Ids: ids})
}

func (a *App) UpdatePerson(ctx context.Context, tx pgx.Tx, orgID, eventID, id uuid.UUID, f PeopleFields) (dbgen.Person, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.UpdatePerson(ctx, dbgen.UpdatePersonParams{
		ID: id, EventID: eventID, OrganizationID: orgID,
		Name: f.Name, ImageUrl: f.ImageURL, Age: f.Age, Gender: f.Gender,
	})
}

func (a *App) DeletePerson(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
	n, err := a.queries.DeletePerson(ctx, dbgen.DeletePersonParams{ID: id, EventID: eventID, OrganizationID: orgID})
	return n > 0, err
}

func (a *App) ListSubEventIDs(ctx context.Context, personID uuid.UUID) ([]uuid.UUID, error) {
	return a.queries.ListPersonSubEventIDs(ctx, personID)
}

func (a *App) AddToSubEvent(ctx context.Context, tx pgx.Tx, personID, subEventID uuid.UUID) error {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.AddPersonToSubEvent(ctx, dbgen.AddPersonToSubEventParams{PersonID: personID, SubEventID: subEventID})
}

func (a *App) DeleteSubEventLinks(ctx context.Context, tx pgx.Tx, personID uuid.UUID) error {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.DeletePersonSubEvents(ctx, personID)
}

func (a *App) MarkCardSent(ctx context.Context, personID uuid.UUID) error {
	return a.queries.MarkCardSent(ctx, personID)
}
