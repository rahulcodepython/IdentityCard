package people

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/pkg/postgres"
)

type PeopleFields struct {
    Name     string
    ImageURL *string
    Age      *int16
    Gender   *string
}

type UpsertPersonResult struct {
    ID             uuid.UUID `json:"id"`
    OrganizationID uuid.UUID `json:"organization_id"`
    EventID        uuid.UUID `json:"event_id"`
    Name           string    `json:"name"`
    Email          string    `json:"email"`
    Mobile         string    `json:"mobile"`
    ImageURL       *string   `json:"image_url"`
    Age            *int16    `json:"age"`
    Gender         *string   `json:"gender"`
    JoinedAt       *string   `json:"joined_at"`
    CardSentAt     *string   `json:"card_sent_at"`
    Inserted       bool      `json:"inserted"`
}

func (a *App) UpsertPerson(ctx context.Context, tx pgx.Tx, orgID, eventID uuid.UUID, email, mobile string, f PeopleFields, joinedAt *time.Time) (*UpsertPersonResult, error) {
    if tx != nil {
        return postgres.QueryJSONTx[UpsertPersonResult](ctx, tx, UpsertPersonQuery,
            orgID, eventID, email, mobile, f.Name, f.ImageURL, f.Age, f.Gender, joinedAt)
    }
    return postgres.QueryJSON[UpsertPersonResult](ctx, a.pool, UpsertPersonQuery,
        orgID, eventID, email, mobile, f.Name, f.ImageURL, f.Age, f.Gender, joinedAt)
}

func (a *App) GetPerson(ctx context.Context, orgID, eventID, id uuid.UUID) (*PersonResponse, error) {
    return postgres.QueryJSON[PersonResponse](ctx, a.pool, GetPersonQuery, id, eventID, orgID)
}

func (a *App) ListPeople(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID, search string) ([]PersonResponse, error) {
    filter := postgres.NewFilter(eventID, orgID)
    filter.AddRaw("p.event_id = $1 AND p.organization_id = $2")

    if subEventID != nil {
        filter.Add("EXISTS (SELECT 1 FROM people_sub_events pse WHERE pse.person_id = p.id AND pse.sub_event_id = $%d)", *subEventID)
    }
    if search != "" {
        filter.Add3("(p.name ILIKE $%d OR p.email ILIKE $%d OR p.mobile ILIKE $%d)", "%"+search+"%")
    }

    sqlQuery := BuildListPeopleQuery(filter.Join(""))
    return postgres.QueryJSONSlice[PersonResponse](ctx, a.pool, sqlQuery, filter.Args...)
}

func (a *App) ListPeopleByIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) ([]PersonResponse, error) {
    return postgres.QueryJSONSlice[PersonResponse](ctx, a.pool, ListPeopleByIDsQuery, eventID, orgID, ids)
}

func (a *App) UpdatePerson(ctx context.Context, tx pgx.Tx, orgID, eventID, id uuid.UUID, f PeopleFields) (*PersonResponse, error) {
    if tx != nil {
        return postgres.QueryJSONTx[PersonResponse](ctx, tx, UpdatePersonQuery, id, eventID, orgID, f.Name, f.ImageURL, f.Age, f.Gender)
    }
    return postgres.QueryJSON[PersonResponse](ctx, a.pool, UpdatePersonQuery, id, eventID, orgID, f.Name, f.ImageURL, f.Age, f.Gender)
}

func (a *App) DeletePerson(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
    res, err := a.pool.Exec(ctx, DeletePersonQuery, id, eventID, orgID)
    if err != nil {
        return false, postgres.MapPgError(err)
    }
    return res.RowsAffected() > 0, nil
}

func (a *App) ListSubEventIDs(ctx context.Context, personID uuid.UUID) ([]uuid.UUID, error) {
    return postgres.QueryJSONSlice[uuid.UUID](ctx, a.pool, ListPersonSubEventIDsQuery, personID)
}

func (a *App) AddToSubEvent(ctx context.Context, tx pgx.Tx, personID, subEventID uuid.UUID) error {
    if tx != nil {
        return postgres.ExecTx(ctx, tx, AddPersonToSubEventQuery, personID, subEventID)
    }
    return postgres.Exec(ctx, a.pool, AddPersonToSubEventQuery, personID, subEventID)
}

func (a *App) DeleteSubEventLinks(ctx context.Context, tx pgx.Tx, personID uuid.UUID) error {
    if tx != nil {
        return postgres.ExecTx(ctx, tx, DeletePersonSubEventsQuery, personID)
    }
    return postgres.Exec(ctx, a.pool, DeletePersonSubEventsQuery, personID)
}

func (a *App) MarkCardSent(ctx context.Context, personID uuid.UUID) error {
    return postgres.Exec(ctx, a.pool, MarkCardSentQuery, personID)
}
