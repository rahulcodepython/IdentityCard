package register

import "time"

type PublicEventInfo struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    StartDate string  `json:"start_date"`
    EndDate   string  `json:"end_date"`
    Venue     *string `json:"venue,omitempty"`
    Logo      *string `json:"logo,omitempty"`
    Organizer *string `json:"organizer,omitempty"`
}

type PublicFormInfo struct {
    ID     string        `json:"id"`
    Name   string        `json:"name"`
    Fields []interface{} `json:"fields"`
}

type PublicApplyConfigResponse struct {
    EventFormID       string           `json:"event_form_id"`
    Status            string           `json:"status"` // "waiting" | "live"
    IsExpired         bool             `json:"is_expired"`
    IsFull            bool             `json:"is_full"`
    MaxApplicants     int              `json:"max_applicants"`
    CurrentApplicants int              `json:"current_applicants"`
    ExpiresAt         time.Time        `json:"expires_at"`
    Event             *PublicEventInfo `json:"event"`
    Form              *PublicFormInfo  `json:"form"`
}

type SubmitApplicationRequest struct {
    Name  string                 `json:"name" validate:"required"`
    Email string                 `json:"email" validate:"required,email"`
    Phone string                 `json:"phone"`
    Data  map[string]interface{} `json:"data"`
}

type SubmitApplicationResponse struct {
    UserID    string `json:"user_id"`
    Status    string `json:"status"`
    Message   string `json:"message"`
    CreatedAt string `json:"created_at"`
}

type QueuedSubmission struct {
    EventID     string                 `json:"event_id"`
    EventFormID string                 `json:"event_form_id"`
    UserID      string                 `json:"user_id"`
    Name        string                 `json:"name"`
    Email       string                 `json:"email"`
    Phone       string                 `json:"phone"`
    Data        map[string]interface{} `json:"data"`
    CreatedAt   string                 `json:"created_at"`
}
