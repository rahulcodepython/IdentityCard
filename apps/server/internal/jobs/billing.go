package jobs

import (
	"context"
	"time"

	"identitycard-server/internal/services"
)

// NewBillingTasks returns the three periodic sweeps that enforce the
// plan/subscription payment lifecycle:
//  1. flip lapsed monthly/yearly subscriptions to past_due,
//  2. hard-delete events whose funding subscription's grace period has
//     expired (marking that subscription expired),
//  3. retire Flash subscriptions/events once their fixed 1-month
//     post-event retention window has passed.
//
// Each task only ever touches its own domain's table directly
// (services.PlansService for subscriptions, services.EventsService for
// events) — a Flash subscription's event end_date, needed by sweep 3, is
// read through EventsService rather than by PlansService reaching into
// the events table itself, since PlansService can't depend on
// EventsService (EventsService already depends on PlansService; see
// services/plans.service.go's SubscriptionCandidate for the same
// constraint on the create-time path).
func NewBillingTasks(plansService *services.PlansService, eventsService *services.EventsService) []Task {
	return []Task{
		func(ctx context.Context) error { return markPastDue(ctx, plansService) },
		func(ctx context.Context) error { return deleteGraceExpired(ctx, plansService, eventsService) },
		func(ctx context.Context) error { return retireFlash(ctx, plansService, eventsService) },
	}
}

func markPastDue(ctx context.Context, plansService *services.PlansService) error {
	return plansService.SweepPastDue(ctx, time.Now())
}

func deleteGraceExpired(ctx context.Context, plansService *services.PlansService, eventsService *services.EventsService) error {
	subs, err := plansService.ListGraceExpired(ctx, time.Now())
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if err := eventsService.DeleteForSubscription(ctx, sub.ID); err != nil {
			return err
		}
		if err := plansService.MarkExpired(ctx, sub.ID); err != nil {
			return err
		}
	}
	return nil
}

// retireFlash deletes a Flash-funded event once it's more than a month
// past its own end_date, and expires the subscription that funded it.
// Flash subscriptions never go past_due (they're one_time, not recurring
// — see ListPastDueCandidateSubscriptions), so this is the only sweep
// that ever touches them.
func retireFlash(ctx context.Context, plansService *services.PlansService, eventsService *services.EventsService) error {
	subs, err := plansService.ListActiveFlashSubscriptions(ctx)
	if err != nil {
		return err
	}
	cutoff := time.Now().AddDate(0, -1, 0) // retained 1 month past the event's end_date
	for _, sub := range subs {
		endDate, ok, err := eventsService.EndDateForSubscription(ctx, sub.ID)
		if err != nil {
			return err
		}
		if !ok || endDate.After(cutoff) {
			continue // no event yet, or still within the retention window
		}
		if err := eventsService.DeleteForSubscription(ctx, sub.ID); err != nil {
			return err
		}
		if err := plansService.MarkExpired(ctx, sub.ID); err != nil {
			return err
		}
	}
	return nil
}
