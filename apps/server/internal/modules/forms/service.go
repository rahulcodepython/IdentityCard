package forms

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/events"
	"identitycard-server/internal/modules/people"
	"identitycard-server/internal/modules/subevents"
)

type Service struct {
	repo      *Repository
	events    *events.Service
	subevents *subevents.Service
	people    *people.Service
}

func NewService(repo *Repository, eventsService *events.Service, subEventsService *subevents.Service, peopleService *people.Service) *Service {
	return &Service{repo: repo, events: eventsService, subevents: subEventsService, people: peopleService}
}

func (s *Service) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreateFormRequest) (FormResponse, error) {
	if _, err := s.events.GetContext(ctx, orgID, eventID); err != nil {
		return FormResponse{}, err
	}

	subEventID := pgtype.UUID{}
	if req.SubEventID != nil {
		if _, err := s.subevents.Get(ctx, orgID, eventID, *req.SubEventID); err != nil {
			return FormResponse{}, httpx.ErrValidation(map[string]string{"sub_event_id": "unknown sub-event"})
		}
		subEventID = pgtype.UUID{Bytes: [16]byte(*req.SubEventID), Valid: true}
	}

	token, err := generateToken()
	if err != nil {
		return FormResponse{}, httpx.ErrInternal()
	}

	form, err := s.repo.Create(ctx, orgID, eventID, subEventID, token, toPgInt4(req.Capacity))
	if err != nil {
		return FormResponse{}, httpx.ErrInternal()
	}
	return toFormResponse(form), nil
}

func (s *Service) List(ctx context.Context, orgID, eventID uuid.UUID) ([]FormResponse, error) {
	if _, err := s.events.GetContext(ctx, orgID, eventID); err != nil {
		return nil, err
	}
	rows, err := s.repo.List(ctx, orgID, eventID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	resp := make([]FormResponse, len(rows))
	for i, row := range rows {
		resp[i] = toFormResponse(row)
	}
	return resp, nil
}

func (s *Service) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdateFormRequest) (FormResponse, error) {
	form, err := s.repo.Update(ctx, orgID, eventID, id, toPgInt4(req.Capacity), req.IsActive)
	if err != nil {
		return FormResponse{}, httpx.ErrNotFound("form")
	}
	return toFormResponse(form), nil
}

func (s *Service) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	deleted, err := s.repo.Delete(ctx, orgID, eventID, id)
	if err != nil {
		return httpx.ErrInternal()
	}
	if !deleted {
		return httpx.ErrNotFound("form")
	}
	return nil
}

func (s *Service) GetPublic(ctx context.Context, token string) (PublicFormResponse, error) {
	row, err := s.repo.GetPublicByToken(ctx, token)
	if err != nil {
		return PublicFormResponse{}, httpx.ErrNotFound("form")
	}

	var subEventName *string
	if row.SubEventName.Valid {
		subEventName = &row.SubEventName.String
	}
	isOpen := row.IsActive && (!row.Capacity.Valid || row.SubmissionsCount < row.Capacity.Int32)

	return PublicFormResponse{EventName: row.EventName, SubEventName: subEventName, IsOpen: isOpen}, nil
}

// Submit speculatively charges one capacity slot before creating the
// person, then undoes the charge if it turns out this was an existing
// registrant resubmitting (not a new signup) or the upsert failed —
// see people.Service.upsert and event_forms.submissions_count. This
// ordering never lets a person be created without a reserved slot, at the
// cost of a brief, self-correcting overcount under concurrent duplicate
// submissions.
func (s *Service) Submit(ctx context.Context, token string, req SubmitFormRequest) error {
	row, err := s.repo.GetPublicByToken(ctx, token)
	if err != nil {
		return httpx.ErrNotFound("form")
	}

	if _, err := s.repo.IncrementSubmissions(ctx, row.ID); err != nil {
		return httpx.NewError(http.StatusConflict, "form_closed", "this form is no longer accepting submissions")
	}

	var subEventID *uuid.UUID
	if row.SubEventID.Valid {
		id := uuid.UUID(row.SubEventID.Bytes)
		subEventID = &id
	}

	_, inserted, err := s.people.Submit(ctx, row.OrganizationID, row.EventID, people.SubmitInput{
		Email: req.Email, Mobile: req.Mobile, Name: req.Name,
		ImageURL: req.ImageURL, Age: req.Age, Gender: req.Gender,
		SubEventID: subEventID,
	})
	if err != nil {
		_ = s.repo.DecrementSubmissions(ctx, row.ID)
		return err
	}
	if !inserted {
		_ = s.repo.DecrementSubmissions(ctx, row.ID)
	}
	return nil
}

func generateToken() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func toPgInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func toFormResponse(f dbgen.EventForm) FormResponse {
	var subEventID *uuid.UUID
	if f.SubEventID.Valid {
		id := uuid.UUID(f.SubEventID.Bytes)
		subEventID = &id
	}
	var capacity *int32
	if f.Capacity.Valid {
		capacity = &f.Capacity.Int32
	}
	return FormResponse{
		ID: f.ID, Token: f.Token, SubEventID: subEventID,
		Capacity: capacity, SubmissionsCount: f.SubmissionsCount, IsActive: f.IsActive,
	}
}
