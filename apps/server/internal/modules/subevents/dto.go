package subevents

import (
	"github.com/google/uuid"

	"identitycard-server/internal/modules/events"
)

// Sub-events get two of events' four schedule modes — fixed_range and
// selective — never recurring or flash: a sub-event's whole purpose is
// splitting people across a subset of the parent's already-materialized
// days, so it doesn't need its own open-ended recurrence rule, and flash
// (single-day) events don't get sub-events at all in practice.
//
// Days reuses events.EventDayInput/EventDayResponse — same wire shape,
// same date/time formats. An empty Days list means the sub-event runs on
// every day of the parent event, using the parent's own times.

type CreateSubEventRequest struct {
	Name           string                 `json:"name" validate:"required,min=2,max=200"`
	ScheduleMode   string                 `json:"schedule_mode" validate:"required,oneof=fixed_range selective"`
	Days           []events.EventDayInput `json:"days" validate:"omitempty,dive"`
	RangeStart     string                 `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd       string                 `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
	RangeEntryTime string                 `json:"range_entry_time" validate:"omitempty,datetime=15:04"`
	RangeExitTime  string                 `json:"range_exit_time" validate:"omitempty,datetime=15:04"`
}

// UpdateSubEventRequest carries the same fields, minus ScheduleMode —
// like the parent event, a sub-event's schedule mode can't change after
// creation.
type UpdateSubEventRequest struct {
	Name           string                 `json:"name" validate:"required,min=2,max=200"`
	Days           []events.EventDayInput `json:"days" validate:"omitempty,dive"`
	RangeStart     string                 `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd       string                 `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
	RangeEntryTime string                 `json:"range_entry_time" validate:"omitempty,datetime=15:04"`
	RangeExitTime  string                 `json:"range_exit_time" validate:"omitempty,datetime=15:04"`
}

type SubEventResponse struct {
	ID           uuid.UUID                 `json:"id"`
	Name         string                    `json:"name"`
	ScheduleMode string                    `json:"schedule_mode"`
	Days         []events.EventDayResponse `json:"days"`
}
