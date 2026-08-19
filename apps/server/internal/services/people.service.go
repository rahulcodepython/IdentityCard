package services

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
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/entities"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/repositories"
	"identitycard-server/internal/utils"
)

type PeopleService struct {
	repo      *repositories.PeopleRepository
	events    *EventsService
	subevents *SubEventsService
	pool      *pgxpool.Pool
}

func NewPeopleService(repo *repositories.PeopleRepository, eventsService *EventsService, subEventsService *SubEventsService, pool *pgxpool.Pool) *PeopleService {
	return &PeopleService{repo: repo, events: eventsService, subevents: subEventsService, pool: pool}
}

func (s *PeopleService) Create(ctx context.Context, orgID, eventID uuid.UUID, req entities.CreatePersonRequest) (entities.PersonResponse, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return entities.PersonResponse{}, err
	}

	resp, _, err := s.upsert(ctx, orgID, eventID, peopleUpsertRequest{
		Email: req.Email, Mobile: req.Mobile, Name: req.Name,
		ImageURL: req.ImageURL, Age: req.Age, Gender: req.Gender,
		SubEventIDs: req.SubEventIDs,
	}, !parent.IsDraft())
	return resp, err
}

// PeopleSubmitInput is the public-form entrypoint used by FormsService —
// unauthenticated, scoped to at most one sub-event (the form's own), and
// already capacity-checked by the caller before this runs.
type PeopleSubmitInput struct {
	Email      string
	Mobile     string
	Name       string
	ImageURL   string
	Age        *int16
	Gender     string
	SubEventID *uuid.UUID
}

// Submit reports whether the person was newly created (vs. an existing
// registrant resubmitting) so FormsService can undo a speculative capacity
// charge on a resubmission — see FormsService.Submit.
func (s *PeopleService) Submit(ctx context.Context, orgID, eventID uuid.UUID, in PeopleSubmitInput) (entities.PersonResponse, bool, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return entities.PersonResponse{}, false, err
	}

	var subEventIDs []uuid.UUID
	if in.SubEventID != nil {
		subEventIDs = []uuid.UUID{*in.SubEventID}
	}

	return s.upsert(ctx, orgID, eventID, peopleUpsertRequest{
		Email: in.Email, Mobile: in.Mobile, Name: in.Name,
		ImageURL: in.ImageURL, Age: in.Age, Gender: in.Gender,
		SubEventIDs: subEventIDs,
	}, !parent.IsDraft())
}

func (s *PeopleService) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (entities.PersonResponse, error) {
	person, err := s.repo.Get(ctx, orgID, eventID, id)
	if err != nil {
		return entities.PersonResponse{}, utils.ErrNotFound("person")
	}
	subEventIDs, err := s.repo.ListSubEventIDs(ctx, id)
	if err != nil {
		return entities.PersonResponse{}, utils.ErrInternal()
	}
	return toPersonResponse(person, subEventIDs), nil
}

// MarkCardSent is called by CardsService after a successful email send —
// it's the only write PeopleService exposes that doesn't go through the
// upsert core, since it's not attendee data, just delivery bookkeeping.
func (s *PeopleService) MarkCardSent(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.MarkCardSent(ctx, id); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

type PeopleListFilter struct {
	SubEventID *uuid.UUID
	Search     *string
}

// List (and Export below) do one sub-event-id query per person rather
// than a join — fine at the scale this is built for; worth revisiting if
// event rosters grow into the thousands.
func (s *PeopleService) List(ctx context.Context, orgID, eventID uuid.UUID, filter PeopleListFilter) ([]entities.PersonResponse, error) {
	rows, err := s.repo.List(ctx, orgID, eventID, peopleToPgUUID(filter.SubEventID), peopleToPgText(peopleDerefOrEmpty(filter.Search)))
	if err != nil {
		return nil, utils.ErrInternal()
	}

	resp := make([]entities.PersonResponse, len(rows))
	for i, row := range rows {
		subEventIDs, err := s.repo.ListSubEventIDs(ctx, row.ID)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		resp[i] = toPersonResponse(row, subEventIDs)
	}
	return resp, nil
}

func (s *PeopleService) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req entities.UpdatePersonRequest) (entities.PersonResponse, error) {
	if err := s.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
		return entities.PersonResponse{}, err
	}

	fields := repositories.PeopleFields{
		Name:     req.Name,
		ImageURL: peopleToPgText(req.ImageURL),
		Age:      peopleToPgInt2(req.Age),
		Gender:   peopleToPgText(req.Gender),
	}

	var person dbgen.Person
	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		updated, err := repo.Update(ctx, orgID, eventID, id, fields)
		if err != nil {
			return utils.ErrNotFound("person")
		}
		person = updated

		if err := repo.DeleteSubEventLinks(ctx, id); err != nil {
			return utils.ErrInternal()
		}
		for _, subEventID := range req.SubEventIDs {
			if err := repo.AddToSubEvent(ctx, id, subEventID); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return entities.PersonResponse{}, txErr
	}

	return toPersonResponse(person, req.SubEventIDs), nil
}

func (s *PeopleService) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	deleted, err := s.repo.Delete(ctx, orgID, eventID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !deleted {
		return utils.ErrNotFound("person")
	}
	return nil
}

type PeopleImportOptions struct {
	SubEventID *uuid.UUID
}

var peopleRequiredCSVColumns = []string{"email", "mobile", "name"}

// ImportCSV parses-and-discards: rows are extracted and turned into people
// records (via the same upsert core as Create/Submit); the uploaded file
// itself is never persisted. Each row is its own upsert — one bad row is
// skipped and reported rather than failing the whole file.
func (s *PeopleService) ImportCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader, opts PeopleImportOptions) (entities.PeopleImportSummary, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return entities.PeopleImportSummary{}, err
	}
	if opts.SubEventID != nil {
		if err := s.validateSubEventIDs(ctx, orgID, eventID, []uuid.UUID{*opts.SubEventID}); err != nil {
			return entities.PeopleImportSummary{}, err
		}
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return entities.PeopleImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "could not read CSV header")
	}
	columns := make(map[string]int, len(header))
	for i, col := range header {
		columns[strings.ToLower(strings.TrimSpace(col))] = i
	}
	for _, required := range peopleRequiredCSVColumns {
		if _, ok := columns[required]; !ok {
			return entities.PeopleImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "missing required column: "+required)
		}
	}

	var subEventIDs []uuid.UUID
	if opts.SubEventID != nil {
		subEventIDs = []uuid.UUID{*opts.SubEventID}
	}

	summary := entities.PeopleImportSummary{}
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, entities.PeopleImportRowError{Row: rowNum, Message: "malformed row"})
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
			summary.Errors = append(summary.Errors, entities.PeopleImportRowError{Row: rowNum, Message: "missing email, mobile, or name"})
			continue
		}
		if _, err := mail.ParseAddress(email); err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, entities.PeopleImportRowError{Row: rowNum, Message: "invalid email"})
			continue
		}

		var age *int16
		if raw := field("age"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 16)
			if err != nil || parsed < 0 || parsed > 150 {
				summary.Skipped++
				summary.Errors = append(summary.Errors, entities.PeopleImportRowError{Row: rowNum, Message: "invalid age"})
				continue
			}
			v := int16(parsed)
			age = &v
		}

		_, inserted, err := s.upsert(ctx, orgID, eventID, peopleUpsertRequest{
			Email: email, Mobile: mobile, Name: name,
			ImageURL: field("image_url"), Age: age, Gender: field("gender"),
			SubEventIDs: subEventIDs,
		}, !parent.IsDraft())
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, entities.PeopleImportRowError{Row: rowNum, Message: "could not save row"})
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

// Export returns a CSV file, not the usual JSON envelope — a deliberate,
// narrow exception (see PeopleController) for a file-download response.
func (s *PeopleService) Export(ctx context.Context, orgID, eventID uuid.UUID, filter PeopleListFilter, ids []uuid.UUID) ([]byte, error) {
	var rows []dbgen.Person
	var err error
	if len(ids) > 0 {
		rows, err = s.repo.ListByIDs(ctx, orgID, eventID, ids)
	} else {
		rows, err = s.repo.List(ctx, orgID, eventID, peopleToPgUUID(filter.SubEventID), peopleToPgText(peopleDerefOrEmpty(filter.Search)))
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

// upsert is the one place that writes a person row — Create, Submit, and
// ImportCSV all funnel through it. Returns whether the row was newly
// created (vs. an existing person resubmitting/being re-imported).
func (s *PeopleService) upsert(ctx context.Context, orgID, eventID uuid.UUID, req peopleUpsertRequest, eventPublished bool) (entities.PersonResponse, bool, error) {
	if err := s.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
		return entities.PersonResponse{}, false, err
	}

	fields := repositories.PeopleFields{
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
	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		upserted, err := repo.Upsert(ctx, orgID, eventID, req.Email, req.Mobile, fields, joinedAt)
		if err != nil {
			return utils.ErrInternal()
		}
		row = upserted

		for _, subEventID := range req.SubEventIDs {
			if err := repo.AddToSubEvent(ctx, row.ID, subEventID); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return entities.PersonResponse{}, false, txErr
	}

	subEventIDs, err := s.repo.ListSubEventIDs(ctx, row.ID)
	if err != nil {
		return entities.PersonResponse{}, false, utils.ErrInternal()
	}

	return toPersonResponse(personFromUpsertRow(row), subEventIDs), row.Inserted, nil
}

func (s *PeopleService) validateSubEventIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	valid, err := s.subevents.ValidateIDs(ctx, orgID, eventID, ids)
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

func toPersonResponse(p dbgen.Person, subEventIDs []uuid.UUID) entities.PersonResponse {
	return entities.PersonResponse{
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
