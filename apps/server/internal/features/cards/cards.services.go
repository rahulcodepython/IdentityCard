package cards

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/qrtoken"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

func (a *App) GenerateForPerson(ctx context.Context, orgID, eventID, personID uuid.UUID) ([]byte, error) {
	pdfBytes, _, err := a.generate(ctx, orgID, eventID, personID)
	return pdfBytes, err
}

func (a *App) ResendForPerson(ctx context.Context, orgID, eventID, personID uuid.UUID) error {
	event, err := a.events.Get(ctx, orgID, eventID)
	if err != nil {
		return err
	}
	if event.Status != "published" {
		return utils.ErrConflict("cards can only be sent once the event is published")
	}

	pdfBytes, person, err := a.generate(ctx, orgID, eventID, personID)
	if err != nil {
		return err
	}
	if err := a.sendCard(person, event, pdfBytes); err != nil {
		return utils.ErrInternal()
	}
	return a.people.UpdateCardSentAt(ctx, personID)
}

func (a *App) SendForEvent(ctx context.Context, orgID, eventID uuid.UUID) {
	event, err := a.events.Get(ctx, orgID, eventID)
	if err != nil {
		log.Printf("cards: SendForEvent: get event %s: %v", eventID, err)
		return
	}

	attendees, err := a.people.List(ctx, orgID, eventID, people.PeopleListFilter{})
	if err != nil {
		log.Printf("cards: SendForEvent: list people for event %s: %v", eventID, err)
		return
	}

	for _, person := range attendees {
		pdfBytes, _, err := a.generate(ctx, orgID, eventID, person.ID)
		if err != nil {
			log.Printf("cards: SendForEvent: generate for person %s: %v", person.ID, err)
			continue
		}
		if err := a.sendCard(person, event, pdfBytes); err != nil {
			log.Printf("cards: SendForEvent: send to %s: %v", person.Email, err)
			continue
		}
		if err := a.people.UpdateCardSentAt(ctx, person.ID); err != nil {
			log.Printf("cards: SendForEvent: mark sent for person %s: %v", person.ID, err)
		}
	}
}

func (a *App) generate(ctx context.Context, orgID, eventID, personID uuid.UUID) ([]byte, people.PersonResponse, error) {
	event, err := a.events.Get(ctx, orgID, eventID)
	if err != nil {
		return nil, people.PersonResponse{}, err
	}
	person, err := a.people.Get(ctx, orgID, eventID, personID)
	if err != nil {
		return nil, people.PersonResponse{}, err
	}

	var subSchedules []subEventSchedule
	for _, subEventID := range person.SubEventIDs {
		se, err := a.subevents.Get(ctx, orgID, eventID, subEventID)
		if err != nil {
			continue
		}
		subSchedules = append(subSchedules, subEventSchedule{
			Name: se.Name,
			Days: []events.EventDayResponse{{Date: se.Date, EntryTime: se.EntryTime, ExitTime: se.ExitTime}},
		})
	}

	org, err := a.orgs.GetSettings(ctx, orgID)
	if err != nil {
		return nil, people.PersonResponse{}, err
	}
	var orgLogo []byte
	if org.HasLogo {
		if data, _, err := a.orgs.GetLogo(ctx, orgID); err == nil {
			orgLogo = data
		}
	}

	var eventImage []byte
	if event.HasImage {
		if data, _, err := a.events.GetImage(ctx, orgID, eventID); err == nil {
			eventImage = data
		}
	}
	var organizerSignature []byte
	if event.HasOrganizerSignature {
		if data, _, err := a.events.GetOrganizerSignature(ctx, orgID, eventID); err == nil {
			organizerSignature = data
		}
	}

	token, err := qrtoken.Issue(a.cfg.QRSecret, personID, eventID, orgID, qrExpiry(event.EndDate))
	if err != nil {
		return nil, people.PersonResponse{}, utils.ErrInternal()
	}

	photo := fetchPersonPhoto(ctx, peopleDerefOrEmpty(person.ImageURL))

	pdfBytes, err := renderCardPDF(renderInput{
		OrgName: org.Name, OrgLogo: orgLogo,
		Event: event, Person: person, SubEvents: subSchedules,
		PersonPhoto: photo, QRToken: token,
		EventImage: eventImage, OrganizerSignature: organizerSignature,
	})
	if err != nil {
		return nil, people.PersonResponse{}, utils.ErrInternal()
	}
	return pdfBytes, person, nil
}

func (a *App) sendCard(person people.PersonResponse, event events.EventResponse, pdfBytes []byte) error {
	subject := fmt.Sprintf("Your ID card for %s", event.Name)
	body := fmt.Sprintf(
		"Hi %s,\n\nYour ID card for %s is attached as a PDF. Please bring it (digitally or printed) to the event.\n",
		person.Name, event.Name,
	)
	return a.mailer.Send(person.Email, subject, body, mailer.Attachment{
		Filename: "id-card.pdf", ContentType: "application/pdf", Data: pdfBytes,
	})
}

func qrExpiry(endDate string) time.Time {
	if endDate == "" {
		return time.Now().Add(48 * time.Hour)
	}
	t, err := time.Parse(timeutil.DateLayout, endDate)
	if err != nil {
		return time.Now().Add(24 * time.Hour)
	}
	return t.Add(48 * time.Hour)
}

func peopleDerefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
