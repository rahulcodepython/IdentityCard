package attendance

import (
    "time"

    "github.com/google/uuid"

    "identitycard-server/internal/features/people"
)

type AttendanceRecordDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    EventID        uuid.UUID  `json:"event_id"`
    PersonID       uuid.UUID  `json:"person_id"`
    Date           time.Time  `json:"date"`
    EntryAt        *time.Time `json:"entry_at"`
    EntryStatus    *string    `json:"entry_status"`
    EntryDeviceID  *uuid.UUID `json:"entry_device_id"`
    ExitAt         *time.Time `json:"exit_at"`
    ExitStatus     *string    `json:"exit_status"`
    ExitDeviceID   *uuid.UUID `json:"exit_device_id"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

type AttendanceRowDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    EventID        uuid.UUID  `json:"event_id"`
    PersonID       uuid.UUID  `json:"person_id"`
    Date           string     `json:"date"`
    EntryAt        *string    `json:"entry_at"`
    EntryStatus    *string    `json:"entry_status"`
    EntryDeviceID  *uuid.UUID `json:"entry_device_id"`
    ExitAt         *string    `json:"exit_at"`
    ExitStatus     *string    `json:"exit_status"`
    ExitDeviceID   *uuid.UUID `json:"exit_device_id"`
    PersonName     string     `json:"person_name"`
}

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
