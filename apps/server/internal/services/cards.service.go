// cards generates the ID-card PDF (with an embedded QR code) for an
// event's attendees, on demand — never stored, per the project's
// data-retention decisions — and emails it. It reads events, people,
// subevents, and organizations through their Services (never their
// repositories) and is itself a one-directional leaf: nothing else in
// this package imports it back in, so EventsController reaches it through
// a small interface (controllers.EventsCardSender) rather than events
// importing it, which would cycle.
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"identitycard-server/internal/config"
	"identitycard-server/internal/entities"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/qrtoken"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

type CardsService struct {
	cfg       *config.Config
	events    *EventsService
	people    *PeopleService
	subevents *SubEventsService
	orgs      *OrganizationsService
	mailer    *mailer.Mailer
}

func NewCardsService(
	cfg *config.Config,
	eventsService *EventsService,
	peopleService *PeopleService,
	subEventsService *SubEventsService,
	orgsService *OrganizationsService,
	m *mailer.Mailer,
) *CardsService {
	return &CardsService{
		cfg: cfg, events: eventsService, people: peopleService,
		subevents: subEventsService, orgs: orgsService, mailer: m,
	}
}

// GenerateForPerson is the on-demand download path — works regardless of
// the event's publish status, so an admin can preview a card before
// publishing.
func (s *CardsService) GenerateForPerson(ctx context.Context, orgID, eventID, personID uuid.UUID) ([]byte, error) {
	pdfBytes, _, err := s.generate(ctx, orgID, eventID, personID)
	return pdfBytes, err
}

// ResendForPerson requires the event to already be published — a card
// only means something fixed (a settled schedule, a QR expiry tied to the
// event's dates) once the event stops changing under it.
func (s *CardsService) ResendForPerson(ctx context.Context, orgID, eventID, personID uuid.UUID) error {
	event, err := s.events.Get(ctx, orgID, eventID)
	if err != nil {
		return err
	}
	if event.Status != "published" {
		return utils.ErrConflict("cards can only be sent once the event is published")
	}

	pdfBytes, person, err := s.generate(ctx, orgID, eventID, personID)
	if err != nil {
		return err
	}
	if err := s.sendCard(person, event, pdfBytes); err != nil {
		return utils.ErrInternal()
	}
	return s.people.MarkCardSent(ctx, personID)
}

// SendForEvent emails every current attendee of a just-published event
// their card. Best-effort and fire-and-forget by design — its only caller,
// EventsController.Publish, runs it in a goroutine so publishing doesn't
// block on however many emails there are to send. A per-recipient failure
// is logged and skipped, not retried; a durable queue (Redis is already in
// the stack) is the natural hardening step if that's ever not good enough.
func (s *CardsService) SendForEvent(ctx context.Context, orgID, eventID uuid.UUID) {
	event, err := s.events.Get(ctx, orgID, eventID)
	if err != nil {
		log.Printf("cards: SendForEvent: get event %s: %v", eventID, err)
		return
	}

	attendees, err := s.people.List(ctx, orgID, eventID, PeopleListFilter{})
	if err != nil {
		log.Printf("cards: SendForEvent: list people for event %s: %v", eventID, err)
		return
	}

	for _, person := range attendees {
		pdfBytes, _, err := s.generate(ctx, orgID, eventID, person.ID)
		if err != nil {
			log.Printf("cards: SendForEvent: generate for person %s: %v", person.ID, err)
			continue
		}
		if err := s.sendCard(person, event, pdfBytes); err != nil {
			log.Printf("cards: SendForEvent: send to %s: %v", person.Email, err)
			continue
		}
		if err := s.people.MarkCardSent(ctx, person.ID); err != nil {
			log.Printf("cards: SendForEvent: mark sent for person %s: %v", person.ID, err)
		}
	}
}

func (s *CardsService) generate(ctx context.Context, orgID, eventID, personID uuid.UUID) ([]byte, entities.PersonResponse, error) {
	event, err := s.events.Get(ctx, orgID, eventID)
	if err != nil {
		return nil, entities.PersonResponse{}, err
	}
	person, err := s.people.Get(ctx, orgID, eventID, personID)
	if err != nil {
		return nil, entities.PersonResponse{}, err
	}

	var subSchedules []subEventSchedule
	for _, subEventID := range person.SubEventIDs {
		se, err := s.subevents.Get(ctx, orgID, eventID, subEventID)
		if err != nil {
			continue // shouldn't happen — a stale link outliving its sub-event; skip rather than fail the whole card
		}
		subSchedules = append(subSchedules, subEventSchedule{Name: se.Name, Days: se.Days})
	}

	org, err := s.orgs.GetSettings(ctx, orgID)
	if err != nil {
		return nil, entities.PersonResponse{}, err
	}
	var orgLogo []byte
	if org.HasLogo {
		if data, _, err := s.orgs.GetLogo(ctx, orgID); err == nil {
			orgLogo = data
		}
	}

	token, err := qrtoken.Issue(s.cfg.QRSecret, personID, eventID, orgID, qrExpiry(event.EndDate))
	if err != nil {
		return nil, entities.PersonResponse{}, utils.ErrInternal()
	}

	photo := fetchPersonPhoto(ctx, peopleDerefOrEmpty(person.ImageURL))

	pdfBytes, err := renderCardPDF(renderInput{
		OrgName: org.Name, OrgLogo: orgLogo,
		Event: event, Person: person, SubEvents: subSchedules,
		PersonPhoto: photo, QRToken: token,
	})
	if err != nil {
		return nil, entities.PersonResponse{}, utils.ErrInternal()
	}
	return pdfBytes, person, nil
}

func (s *CardsService) sendCard(person entities.PersonResponse, event entities.EventResponse, pdfBytes []byte) error {
	subject := fmt.Sprintf("Your ID card for %s", event.Name)
	body := fmt.Sprintf(
		"Hi %s,\n\nYour ID card for %s is attached as a PDF. Please bring it (digitally or printed) to the event.\n",
		person.Name, event.Name,
	)
	return s.mailer.Send(person.Email, subject, body, mailer.Attachment{
		Filename: "id-card.pdf", ContentType: "application/pdf", Data: pdfBytes,
	})
}

// qrExpiry gives a card's QR token a validity window past the event's
// last day. An open-ended recurring event (endDate nil — see
// entities.EventSummary.EndDate) has no last day to key off, so it gets a
// flat 1-year expiry instead, matching how far ahead its schedule is kept
// materialized (see EventsService's recurringHorizonDays) — the card just
// needs reissuing whenever that horizon is renewed.
func qrExpiry(endDate *string) time.Time {
	if endDate == nil {
		return time.Now().AddDate(1, 0, 0)
	}
	t, err := time.Parse(timeutil.DateLayout, *endDate)
	if err != nil {
		return time.Now().Add(24 * time.Hour)
	}
	return t.Add(48 * time.Hour)
}
