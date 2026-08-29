package people

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/utils"
)

func (a *App) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreatePersonRequest) (PersonResponse, error) {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return PersonResponse{}, err
	}

	resp, _, err := a.upsert(ctx, orgID, eventID, peopleUpsertRequest{
		Email: req.Email, Mobile: req.Mobile, Name: req.Name,
		ImageURL: req.ImageURL, Age: req.Age, Gender: req.Gender,
		SubEventIDs: req.SubEventIDs,
	}, !parent.IsDraft())
	return resp, err
}

func (a *App) Submit(ctx context.Context, orgID, eventID uuid.UUID, in PeopleSubmitInput) (PersonResponse, bool, error) {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return PersonResponse{}, false, err
	}

	var subEventIDs []uuid.UUID
	if in.SubEventID != nil {
		subEventIDs = []uuid.UUID{*in.SubEventID}
	}

	return a.upsert(ctx, orgID, eventID, peopleUpsertRequest{
		Email: in.Email, Mobile: in.Mobile, Name: in.Name,
		ImageURL: in.ImageURL, Age: in.Age, Gender: in.Gender,
		SubEventIDs: subEventIDs,
	}, !parent.IsDraft())
}

func (a *App) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (PersonResponse, error) {
	person, err := a.GetPerson(ctx, orgID, eventID, id)
	if err != nil {
		return PersonResponse{}, utils.ErrNotFound("person")
	}
	subEventIDs, err := a.ListSubEventIDs(ctx, id)
	if err != nil {
		return PersonResponse{}, utils.ErrInternal()
	}
	return toPersonResponse(person, subEventIDs), nil
}

func (a *App) UpdateCardSentAt(ctx context.Context, id uuid.UUID) error {
	if err := a.MarkCardSent(ctx, id); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

func (a *App) List(ctx context.Context, orgID, eventID uuid.UUID, filter PeopleListFilter) ([]PersonResponse, error) {
	rows, err := a.ListPeople(ctx, orgID, eventID, peopleToPgUUID(filter.SubEventID), peopleToPgText(peopleDerefOrEmpty(filter.Search)))
	if err != nil {
		return nil, utils.ErrInternal()
	}

	resp := make([]PersonResponse, len(rows))
	for i, row := range rows {
		subEventIDs, err := a.ListSubEventIDs(ctx, row.ID)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		resp[i] = toPersonResponse(row, subEventIDs)
	}
	return resp, nil
}

func (a *App) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdatePersonRequest) (PersonResponse, error) {
	if err := a.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
		return PersonResponse{}, err
	}

	fields := PeopleFields{
		Name:     req.Name,
		ImageURL: peopleToPgText(req.ImageURL),
		Age:      peopleToPgInt2(req.Age),
		Gender:   peopleToPgText(req.Gender),
	}

	var person dbgen.Person
	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		updated, err := a.UpdatePerson(ctx, tx, orgID, eventID, id, fields)
		if err != nil {
			return utils.ErrNotFound("person")
		}
		person = updated

		if err := a.DeleteSubEventLinks(ctx, tx, id); err != nil {
			return utils.ErrInternal()
		}
		for _, subEventID := range req.SubEventIDs {
			if err := a.AddToSubEvent(ctx, tx, id, subEventID); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return PersonResponse{}, txErr
	}

	return toPersonResponse(person, req.SubEventIDs), nil
}

func (a *App) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	deleted, err := a.DeletePerson(ctx, orgID, eventID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !deleted {
		return utils.ErrNotFound("person")
	}
	return nil
}

var peopleRequiredCSVColumns = []string{"email", "mobile", "name"}

func (a *App) ImportCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader, opts PeopleImportOptions) (PeopleImportSummary, error) {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return PeopleImportSummary{}, err
	}
	if opts.SubEventID != nil {
		if err := a.validateSubEventIDs(ctx, orgID, eventID, []uuid.UUID{*opts.SubEventID}); err != nil {
			return PeopleImportSummary{}, err
		}
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return PeopleImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "could not read CSV header")
	}
	columns := make(map[string]int, len(header))
	for i, col := range header {
		columns[strings.ToLower(strings.TrimSpace(col))] = i
	}
	for _, required := range peopleRequiredCSVColumns {
		if _, ok := columns[required]; !ok {
			return PeopleImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "missing required column: "+required)
		}
	}

	var subEventIDs []uuid.UUID
	if opts.SubEventID != nil {
		subEventIDs = []uuid.UUID{*opts.SubEventID}
	}

	summary := PeopleImportSummary{}
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "malformed row"})
			continue
		}

		field := func(col string) string {
			idx, ok := columns[col]
			if !ok || idx >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[idx])
		}

		email, mobile, name := field("email"), field("mobile"), field("name")
		if email == "" || mobile == "" || name == "" {
			summary.Skipped++
			summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "missing email, mobile, or name"})
			continue
		}
		if _, err := mail.ParseAddress(email); err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "invalid email"})
			continue
		}

		var age *int16
		if raw := field("age"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 16)
			if err != nil || parsed < 0 || parsed > 150 {
				summary.Skipped++
				summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "invalid age"})
				continue
			}
			v := int16(parsed)
			age = &v
		}

		_, inserted, err := a.upsert(ctx, orgID, eventID, peopleUpsertRequest{
			Email: email, Mobile: mobile, Name: name,
			ImageURL: field("image_url"), Age: age, Gender: field("gender"),
			SubEventIDs: subEventIDs,
		}, !parent.IsDraft())
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "could not save row"})
			continue
		}
		if inserted {
			summary.Inserted++
		} else {
			summary.Updated++
		}
	}

	return summary, nil
}

func (a *App) Export(ctx context.Context, orgID, eventID uuid.UUID, filter PeopleListFilter, ids []uuid.UUID) ([]byte, error) {
	var rows []dbgen.Person
	var err error
	if len(ids) > 0 {
		rows, err = a.ListPeopleByIDs(ctx, orgID, eventID, ids)
	} else {
		rows, err = a.ListPeople(ctx, orgID, eventID, peopleToPgUUID(filter.SubEventID), peopleToPgText(peopleDerefOrEmpty(filter.Search)))
	}
	if err != nil {
		return nil, utils.ErrInternal()
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"email", "mobile", "name", "image_url", "age", "gender", "joined_at"})
	for _, p := range rows {
		_ = w.Write([]string{
			p.Email, p.Mobile, p.Name,
			p.ImageUrl.String, peopleFormatPgInt2(p.Age), p.Gender.String,
			peopleFormatPgTimestamptz(p.JoinedAt),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, utils.ErrInternal()
	}
	return buf.Bytes(), nil
}

type peopleUpsertRequest struct {
	Email       string
	Mobile      string
	Name        string
	ImageURL    string
	Age         *int16
	Gender      string
	SubEventIDs []uuid.UUID
}

func (a *App) upsert(ctx context.Context, orgID, eventID uuid.UUID, req peopleUpsertRequest, eventPublished bool) (PersonResponse, bool, error) {
	if err := a.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
		return PersonResponse{}, false, err
	}

	fields := PeopleFields{
		Name:     req.Name,
		ImageURL: peopleToPgText(req.ImageURL),
		Age:      peopleToPgInt2(req.Age),
		Gender:   peopleToPgText(req.Gender),
	}

	var joinedAt pgtype.Timestamptz
	if eventPublished {
		joinedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	var row dbgen.UpsertPersonRow
	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		upserted, err := a.UpsertPerson(ctx, tx, orgID, eventID, req.Email, req.Mobile, fields, joinedAt)
		if err != nil {
			return utils.ErrInternal()
		}
		row = upserted

		for _, subEventID := range req.SubEventIDs {
			if err := a.AddToSubEvent(ctx, tx, row.ID, subEventID); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return PersonResponse{}, false, txErr
	}

	subEventIDs, err := a.ListSubEventIDs(ctx, row.ID)
	if err != nil {
		return PersonResponse{}, false, utils.ErrInternal()
	}

	return toPersonResponse(personFromUpsertRow(row), subEventIDs), row.Inserted, nil
}

func (a *App) validateSubEventIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	valid, err := a.subevents.ValidateIDs(ctx, orgID, eventID, ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if !valid[id] {
			return utils.ErrValidation(map[string]string{"sub_event_ids": "unknown sub-event: " + id.String()})
		}
	}
	return nil
}

func personFromUpsertRow(row dbgen.UpsertPersonRow) dbgen.Person {
	return dbgen.Person{
		ID: row.ID, OrganizationID: row.OrganizationID, EventID: row.EventID,
		Email: row.Email, Mobile: row.Mobile, Name: row.Name,
		ImageUrl: row.ImageUrl, Age: row.Age, Gender: row.Gender,
		JoinedAt: row.JoinedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		CardSentAt: row.CardSentAt,
	}
}

func toPersonResponse(p dbgen.Person, subEventIDs []uuid.UUID) PersonResponse {
	return PersonResponse{
		ID: p.ID, Email: p.Email, Mobile: p.Mobile, Name: p.Name,
		ImageURL: peoplePgTextPtr(p.ImageUrl), Age: peoplePgInt2Ptr(p.Age), Gender: peoplePgTextPtr(p.Gender),
		JoinedAt:    peoplePgTimestamptzPtr(p.JoinedAt),
		CardSentAt:  peoplePgTimestamptzPtr(p.CardSentAt),
		SubEventIDs: subEventIDs,
	}
}

func peopleToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: [16]byte(*id), Valid: true}
}

func peopleToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func peopleToPgInt2(i *int16) pgtype.Int2 {
	if i == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: *i, Valid: true}
}

func peoplePgTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func peoplePgInt2Ptr(i pgtype.Int2) *int16 {
	if !i.Valid {
		return nil
	}
	return &i.Int16
}

func peoplePgTimestamptzPtr(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func peopleFormatPgInt2(i pgtype.Int2) string {
	if !i.Valid {
		return ""
	}
	return strconv.FormatInt(int64(i.Int16), 10)
}

func peopleFormatPgTimestamptz(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

func peopleDerefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
