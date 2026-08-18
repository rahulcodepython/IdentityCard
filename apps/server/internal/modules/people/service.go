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
	"github.com/jackc/pgx/v5/pgxpool"

	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/events"
	"identitycard-server/internal/modules/subevents"
)

type Service struct {
	repo      *Repository
	events    *events.Service
	subevents *subevents.Service
	pool      *pgxpool.Pool
}

func NewService(repo *Repository, eventsService *events.Service, subEventsService *subevents.Service, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, events: eventsService, subevents: subEventsService, pool: pool}
}

func (s *Service) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreatePersonRequest) (PersonResponse, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return PersonResponse{}, err
	}

	resp, _, err := s.upsert(ctx, orgID, eventID, upsertRequest{
		Email: req.Email, Mobile: req.Mobile, Name: req.Name,
		ImageURL: req.ImageURL, Age: req.Age, Gender: req.Gender,
		SubEventIDs: req.SubEventIDs,
	}, !parent.IsDraft())
	return resp, err
}

// SubmitInput is the public-form entrypoint used by forms.Service —
// unauthenticated, scoped to at most one sub-event (the form's own), and
// already capacity-checked by the caller before this runs.
type SubmitInput struct {
	Email      string
	Mobile     string
	Name       string
	ImageURL   string
	Age        *int16
	Gender     string
	SubEventID *uuid.UUID
}

// Submit reports whether the person was newly created (vs. an existing
// registrant resubmitting) so forms.Service can undo a speculative
// capacity charge on a resubmission — see forms.Service.Submit.
func (s *Service) Submit(ctx context.Context, orgID, eventID uuid.UUID, in SubmitInput) (PersonResponse, bool, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return PersonResponse{}, false, err
	}

	var subEventIDs []uuid.UUID
	if in.SubEventID != nil {
		subEventIDs = []uuid.UUID{*in.SubEventID}
	}

	return s.upsert(ctx, orgID, eventID, upsertRequest{
		Email: in.Email, Mobile: in.Mobile, Name: in.Name,
		ImageURL: in.ImageURL, Age: in.Age, Gender: in.Gender,
		SubEventIDs: subEventIDs,
	}, !parent.IsDraft())
}

func (s *Service) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (PersonResponse, error) {
	person, err := s.repo.Get(ctx, orgID, eventID, id)
	if err != nil {
		return PersonResponse{}, httpx.ErrNotFound("person")
	}
	subEventIDs, err := s.repo.ListSubEventIDs(ctx, id)
	if err != nil {
		return PersonResponse{}, httpx.ErrInternal()
	}
	return toPersonResponse(person, subEventIDs), nil
}

// MarkCardSent is called by cards.Service after a successful email send —
// it's the only write people.Service exposes that doesn't go through the
// upsert core, since it's not attendee data, just delivery bookkeeping.
func (s *Service) MarkCardSent(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.MarkCardSent(ctx, id); err != nil {
		return httpx.ErrInternal()
	}
	return nil
}

type ListFilter struct {
	SubEventID *uuid.UUID
	Search     *string
}

// List (and Export below) do one sub-event-id query per person rather
// than a join — fine at the scale this is built for; worth revisiting if
// event rosters grow into the thousands.
func (s *Service) List(ctx context.Context, orgID, eventID uuid.UUID, filter ListFilter) ([]PersonResponse, error) {
	rows, err := s.repo.List(ctx, orgID, eventID, toPgUUID(filter.SubEventID), toPgText(derefOrEmpty(filter.Search)))
	if err != nil {
		return nil, httpx.ErrInternal()
	}

	resp := make([]PersonResponse, len(rows))
	for i, row := range rows {
		subEventIDs, err := s.repo.ListSubEventIDs(ctx, row.ID)
		if err != nil {
			return nil, httpx.ErrInternal()
		}
		resp[i] = toPersonResponse(row, subEventIDs)
	}
	return resp, nil
}

func (s *Service) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdatePersonRequest) (PersonResponse, error) {
	if err := s.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
		return PersonResponse{}, err
	}

	fields := Fields{
		Name:     req.Name,
		ImageURL: toPgText(req.ImageURL),
		Age:      toPgInt2(req.Age),
		Gender:   toPgText(req.Gender),
	}

	var person dbgen.Person
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		updated, err := repo.Update(ctx, orgID, eventID, id, fields)
		if err != nil {
			return httpx.ErrNotFound("person")
		}
		person = updated

		if err := repo.DeleteSubEventLinks(ctx, id); err != nil {
			return httpx.ErrInternal()
		}
		for _, subEventID := range req.SubEventIDs {
			if err := repo.AddToSubEvent(ctx, id, subEventID); err != nil {
				return httpx.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return PersonResponse{}, txErr
	}

	return toPersonResponse(person, req.SubEventIDs), nil
}

func (s *Service) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	deleted, err := s.repo.Delete(ctx, orgID, eventID, id)
	if err != nil {
		return httpx.ErrInternal()
	}
	if !deleted {
		return httpx.ErrNotFound("person")
	}
	return nil
}

type ImportOptions struct {
	SubEventID *uuid.UUID
}

var requiredCSVColumns = []string{"email", "mobile", "name"}

// ImportCSV parses-and-discards: rows are extracted and turned into people
// records (via the same upsert core as Create/Submit); the uploaded file
// itself is never persisted. Each row is its own upsert — one bad row is
// skipped and reported rather than failing the whole file.
func (s *Service) ImportCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader, opts ImportOptions) (ImportSummary, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return ImportSummary{}, err
	}
	if opts.SubEventID != nil {
		if err := s.validateSubEventIDs(ctx, orgID, eventID, []uuid.UUID{*opts.SubEventID}); err != nil {
			return ImportSummary{}, err
		}
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return ImportSummary{}, httpx.NewError(http.StatusBadRequest, "bad_csv", "could not read CSV header")
	}
	columns := make(map[string]int, len(header))
	for i, col := range header {
		columns[strings.ToLower(strings.TrimSpace(col))] = i
	}
	for _, required := range requiredCSVColumns {
		if _, ok := columns[required]; !ok {
			return ImportSummary{}, httpx.NewError(http.StatusBadRequest, "bad_csv", "missing required column: "+required)
		}
	}

	var subEventIDs []uuid.UUID
	if opts.SubEventID != nil {
		subEventIDs = []uuid.UUID{*opts.SubEventID}
	}

	summary := ImportSummary{}
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, ImportRowError{Row: rowNum, Message: "malformed row"})
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
			summary.Errors = append(summary.Errors, ImportRowError{Row: rowNum, Message: "missing email, mobile, or name"})
			continue
		}
		if _, err := mail.ParseAddress(email); err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, ImportRowError{Row: rowNum, Message: "invalid email"})
			continue
		}

		var age *int16
		if raw := field("age"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 16)
			if err != nil || parsed < 0 || parsed > 150 {
				summary.Skipped++
				summary.Errors = append(summary.Errors, ImportRowError{Row: rowNum, Message: "invalid age"})
				continue
			}
			v := int16(parsed)
			age = &v
		}

		_, inserted, err := s.upsert(ctx, orgID, eventID, upsertRequest{
			Email: email, Mobile: mobile, Name: name,
			ImageURL: field("image_url"), Age: age, Gender: field("gender"),
			SubEventIDs: subEventIDs,
		}, !parent.IsDraft())
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, ImportRowError{Row: rowNum, Message: "could not save row"})
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
// narrow exception (see handler.go) for a file-download response.
func (s *Service) Export(ctx context.Context, orgID, eventID uuid.UUID, filter ListFilter, ids []uuid.UUID) ([]byte, error) {
	var rows []dbgen.Person
	var err error
	if len(ids) > 0 {
		rows, err = s.repo.ListByIDs(ctx, orgID, eventID, ids)
	} else {
		rows, err = s.repo.List(ctx, orgID, eventID, toPgUUID(filter.SubEventID), toPgText(derefOrEmpty(filter.Search)))
	}
	if err != nil {
		return nil, httpx.ErrInternal()
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"email", "mobile", "name", "image_url", "age", "gender", "joined_at"})
	for _, p := range rows {
		_ = w.Write([]string{
			p.Email, p.Mobile, p.Name,
			p.ImageUrl.String, formatPgInt2(p.Age), p.Gender.String,
			formatPgTimestamptz(p.JoinedAt),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, httpx.ErrInternal()
	}
	return buf.Bytes(), nil
}

type upsertRequest struct {
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
func (s *Service) upsert(ctx context.Context, orgID, eventID uuid.UUID, req upsertRequest, eventPublished bool) (PersonResponse, bool, error) {
	if err := s.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
		return PersonResponse{}, false, err
	}

	fields := Fields{
		Name:     req.Name,
		ImageURL: toPgText(req.ImageURL),
		Age:      toPgInt2(req.Age),
		Gender:   toPgText(req.Gender),
	}

	var joinedAt pgtype.Timestamptz
	if eventPublished {
		joinedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	var row dbgen.UpsertPersonRow
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		upserted, err := repo.Upsert(ctx, orgID, eventID, req.Email, req.Mobile, fields, joinedAt)
		if err != nil {
			return httpx.ErrInternal()
		}
		row = upserted

		for _, subEventID := range req.SubEventIDs {
			if err := repo.AddToSubEvent(ctx, row.ID, subEventID); err != nil {
				return httpx.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return PersonResponse{}, false, txErr
	}

	subEventIDs, err := s.repo.ListSubEventIDs(ctx, row.ID)
	if err != nil {
		return PersonResponse{}, false, httpx.ErrInternal()
	}

	return toPersonResponse(personFromUpsertRow(row), subEventIDs), row.Inserted, nil
}

func (s *Service) validateSubEventIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	valid, err := s.subevents.ValidateIDs(ctx, orgID, eventID, ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if !valid[id] {
			return httpx.ErrValidation(map[string]string{"sub_event_ids": "unknown sub-event: " + id.String()})
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
		ImageURL: pgTextPtr(p.ImageUrl), Age: pgInt2Ptr(p.Age), Gender: pgTextPtr(p.Gender),
		JoinedAt:    pgTimestamptzPtr(p.JoinedAt),
		CardSentAt:  pgTimestamptzPtr(p.CardSentAt),
		SubEventIDs: subEventIDs,
	}
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: [16]byte(*id), Valid: true}
}

func toPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func toPgInt2(i *int16) pgtype.Int2 {
	if i == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: *i, Valid: true}
}

func pgTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgInt2Ptr(i pgtype.Int2) *int16 {
	if !i.Valid {
		return nil
	}
	return &i.Int16
}

func pgTimestamptzPtr(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func formatPgInt2(i pgtype.Int2) string {
	if !i.Valid {
		return ""
	}
	return strconv.FormatInt(int64(i.Int16), 10)
}

func formatPgTimestamptz(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
