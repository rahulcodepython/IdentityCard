package eventform

import (
    "encoding/json"
    "time"
)

type EventFormDetails struct {
    ID              string          `json:"id"`
    EventID         string          `json:"event_id"`
    Name            string          `json:"name"`
    Fields          json.RawMessage `json:"fields"`
    IsLocked        bool            `json:"is_locked"`
    LockedAt        *time.Time      `json:"locked_at,omitempty"`
    MaxApplicants   int             `json:"max_applicants"`
    ExpiresAt       time.Time       `json:"expires_at"`
    TotalApplicants int             `json:"total_applicants"`
    CanDelete       bool            `json:"can_delete"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
}

type CreateEventFormRequest struct {
    Source        string          `json:"source"` // "template" | "scratch"
    TemplateID    *string         `json:"template_id,omitempty"`
    Name          string          `json:"name"`
    Fields        json.RawMessage `json:"fields,omitempty"`
    MaxApplicants int             `json:"max_applicants"`
    ExpiresAt     time.Time       `json:"expires_at"`
}

type UpdateEventFormRequest struct {
    Name          *string          `json:"name,omitempty"`
    Fields        *json.RawMessage `json:"fields,omitempty"`
    MaxApplicants *int             `json:"max_applicants,omitempty"`
    ExpiresAt     *time.Time       `json:"expires_at,omitempty"`
}
