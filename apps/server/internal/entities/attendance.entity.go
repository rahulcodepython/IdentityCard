package entities

import "github.com/google/uuid"

type ScanRequest struct {
	QRToken string `json:"qr_token" validate:"required"`
}

// ScanResponse is exactly the "simple card popup" the spec asks for: who
// this is, which event, whether this was their entry or exit scan, first
// scan of the day or not, and early/on_time/late against the schedule.
type ScanResponse struct {
	Direction string         `json:"direction"` // entry | exit | already_completed
	Status    *string        `json:"status"`    // early | on_time | late — absent for already_completed
	Date      string         `json:"date"`
	EventName string         `json:"event_name"`
	Person    PersonResponse `json:"person"`
}

// RosterEntry is one (person, date) pair the person was permitted to
// enter on — present whether or not they actually showed up (Attended
// distinguishes the two), which is what makes this a full roster rather
// than just a list of scans. See services.AttendanceService.BuildRoster.
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
