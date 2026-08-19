package jobs

import (
	"context"

	"identitycard-server/internal/services"
)

// NewRecurrenceTasks returns the one sweep that keeps every published
// recurring event's materialized event_days window topped up — see
// services.EventsService.ExtendRecurringHorizon. Runs on the same daily ticker as
// the billing sweeps (see billing.go); a delayed run just catches each
// event up further next time rather than losing coverage.
func NewRecurrenceTasks(eventsService *services.EventsService) []Task {
	return []Task{
		func(ctx context.Context) error { return eventsService.ExtendRecurringHorizon(ctx) },
	}
}
