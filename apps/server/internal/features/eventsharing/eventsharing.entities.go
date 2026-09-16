package eventsharing

import "time"

type EventFormStatus string

const (
    StatusWaiting EventFormStatus = "waiting"
    StatusLive    EventFormStatus = "live"
)

type EventForm struct {
    ID            string          `json:"id"`
    EventID       string          `json:"event_id"`
    FormID        string          `json:"form_id"`
    MaxApplicants int             `json:"max_applicants"`
    ExpiresAt     time.Time       `json:"expires_at"`
    Status        EventFormStatus `json:"status"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
}

type AssignedFormSummary struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    FieldsCount int    `json:"fields_count"`
    IsPublished bool   `json:"is_published"`
}

type EventSharingResponse struct {
    EventForm       *EventForm           `json:"event_form"`
    AssignedForm    *AssignedFormSummary `json:"assigned_form"`
    TotalApplicants int                  `json:"total_applicants"`
    CanChangeForm   bool                 `json:"can_change_form"`
}

type UpdateEventSharingRequest struct {
    FormID        string          `json:"form_id" validate:"required"`
    MaxApplicants int             `json:"max_applicants"` // -1 for unlimited, or >= 1
    ExpiresAt     time.Time       `json:"expires_at" validate:"required"`
    Status        EventFormStatus `json:"status"`
}

type UpsertSharingResult struct {
    StatusCode string                `json:"status_code"`
    Response   *EventSharingResponse `json:"response"`
}
