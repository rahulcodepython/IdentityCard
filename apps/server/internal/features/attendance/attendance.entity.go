package attendance

import (
	"github.com/google/uuid"

	"identitycard-server/internal/features/people"
)

type ScanRequest struct {
	QRToken string `json:"qr_token" validate:"required"`
}

type ScanResponse struct {
	Direction string                `json:"direction"`
	Status    *string               `json:"status"`
	Date      string                `json:"date"`
	EventName string                `json:"event_name"`
	Person    people.PersonResponse `json:"person"`
}

type RosterEntry struct {
	PersonID    uuid.UUID `json:"person_id"`
	PersonName  string    `json:"person_name"`
	PersonEmail string    `json:"person_email"`
	Date        string    `json:"date"`
	Attended    bool      `json:"attended"`
	EntryAt     *string   `json:"entry_at"`
	EntryStatus *string   `json:"entry_status"`
	ExitAt      *string   `json:"exit_at"`
	ExitStatus  *string   `json:"exit_status"`
}

type RosterFilter struct {
	SubEventID *uuid.UUID
	Date       *string
	Attended   *bool
	Status     *string
}
