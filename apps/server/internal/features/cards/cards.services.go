package cards

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/qrtoken"
	"identitycard-server/internal/utils"
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
		return utils.ErrConflict("Cards can only be sent once the event is published.", generic.ErrEventsNotDraft)
	}

    pdfBytes, person, err := a.generate(ctx, orgID, eventID, personID)
    if err != nil {
        return err
    }
    if err := a.sendCard(ctx, person, event, pdfBytes); err != nil {
        return utils.ErrInternal("Failed to send ID card email.", err)
    }
    return a.people.UpdateCardSentAt(ctx, personID)
}

func (a *App) SendForEvent(ctx context.Context, orgID, eventID uuid.UUID) error {
    event, err := a.events.Get(ctx, orgID, eventID)
    if err != nil {
        log.Printf("cards: SendForEvent: get event %s: %v", eventID, err)
        return err
    }

    attendees, err := a.people.List(ctx, orgID, eventID, people.PeopleListFilter{})
    if err != nil {
        log.Printf("cards: SendForEvent: list people for event %s: %v", eventID, err)
        return err
    }

    var batch []mailer.BatchMessage
    var sentPersonIDs []uuid.UUID

    for _, person := range attendees {
        pdfBytes, _, err := a.generate(ctx, orgID, eventID, person.ID)
        if err != nil {
            log.Printf("cards: SendForEvent: generate for person %s: %v", person.ID, err)
            continue
        }

        subject := fmt.Sprintf("Your ID card for %s", event.Name)
        body := fmt.Sprintf(
            "Hi %s,\n\nYour ID card for %s is attached as a PDF. Please bring it (digitally or printed) to the event.\n",
            person.Name, event.Name,
        )

        batch = append(batch, mailer.BatchMessage{
            To:      person.Email,
            Subject: subject,
            Body:    body,
            Attachments: []mailer.Attachment{
                {Filename: "id-card.pdf", ContentType: "application/pdf", Data: pdfBytes},
            },
        })
        sentPersonIDs = append(sentPersonIDs, person.ID)
    }

    if len(batch) > 0 {
        if err := a.mailer.SendBatch(ctx, batch); err != nil {
            log.Printf("cards: SendForEvent: batch send via Resend failed: %v", err)
            return err
        }
        for _, pid := range sentPersonIDs {
            if err := a.people.UpdateCardSentAt(ctx, pid); err != nil {
                log.Printf("cards: SendForEvent: mark sent for person %s: %v", pid, err)
            }
        }
    }
    return nil
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
		return nil, people.PersonResponse{}, utils.ErrInternal("Failed to issue QR token.", err)
	}

	photo := fetchPersonPhoto(ctx, peopleDerefOrEmpty(person.ImageURL))

	pdfBytes, err := renderCardPDF(renderInput{
		OrgName: org.Name, OrgLogo: orgLogo,
		Event: event, Person: person, SubEvents: subSchedules,
		PersonPhoto: photo, QRToken: token,
		EventImage: eventImage, OrganizerSignature: organizerSignature,
	})
	if err != nil {
		return nil, people.PersonResponse{}, utils.ErrInternal("Failed to render card PDF.", err)
	}
	return pdfBytes, person, nil
}

func (a *App) sendCard(ctx context.Context, person people.PersonResponse, event events.EventResponse, pdfBytes []byte) error {
    subject := fmt.Sprintf("Your ID card for %s", event.Name)
    body := fmt.Sprintf(
        "Hi %s,\n\nYour ID card for %s is attached as a PDF. Please bring it (digitally or printed) to the event.\n",
        person.Name, event.Name,
    )
    return a.mailer.Send(ctx, person.Email, subject, body, mailer.Attachment{
        Filename: "id-card.pdf", ContentType: "application/pdf", Data: pdfBytes,
    })
}

func qrExpiry(endDate string) time.Time {
	if endDate == "" {
		return time.Now().Add(48 * time.Hour)
	}
	t, err := time.Parse(utils.DateLayout, endDate)
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
