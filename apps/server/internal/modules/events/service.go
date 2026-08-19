package events

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/plans"
	"identitycard-server/internal/timeutil"
)

type Service struct {
	repo  *Repository
	pool  *pgxpool.Pool
	plans *plans.Service
}

func NewService(repo *Repository, pool *pgxpool.Pool, plansService *plans.Service) *Service {
	return &Service{repo: repo, pool: pool, plans: plansService}
}

// recurringHorizonDays is how far ahead a recurring event's event_days are
// kept materialized (confirmed with the user) — ExtendRecurringHorizon
// advances this window by another day's worth of headroom each time
// internal/jobs runs it.
const recurringHorizonDays = 365

// fixedRangeMaxSpanDays caps a fixed_range event's [start,end] span —
// truly open-ended schedules belong to recurring instead.
const fixedRangeMaxSpanDays = 365 * 2

// checkEventLimit picks which of the org's subscriptions should fund a
// new event of the given schedule mode, or reports why none can.
// Subscriptions are considered oldest-first (see
// plans.Service.ActiveSubscriptionsForOrgOrderedByAge) so an org spends
// its earliest-purchased capacity before newer top-ups — that FIFO order
// is what keeps the later per-subscription grace/deletion sweep correct.
//
// A flash-kind subscription can only ever fund a schedule_mode='flash'
// event, and conversely a flash event can only ever be funded by a
// flash-kind subscription — confirmed: Flash is a distinct one-time
// product, not something an Unlimited/Custom/Base plan happens to cover.
func (s *Service) checkEventLimit(ctx context.Context, orgID uuid.UUID, scheduleMode string) (uuid.UUID, error) {
	candidates, err := s.plans.ActiveSubscriptionsForOrgOrderedByAge(ctx, orgID)
	if err != nil {
		return uuid.Nil, err
	}

	wantFlash := scheduleMode == "flash"
	var matched bool
	for _, c := range candidates {
		if (c.Kind == "flash") != wantFlash {
			continue
		}
		matched = true
		if c.Kind == "unlimited" {
			return c.SubscriptionID, nil
		}
		if c.EventQuota == nil {
			continue
		}
		count, err := s.repo.CountBySubscription(ctx, c.SubscriptionID)
		if err != nil {
			return uuid.Nil, httpx.ErrInternal()
		}
		if count < int64(*c.EventQuota) {
			return c.SubscriptionID, nil
		}
	}
	if !matched {
		if wantFlash {
			return uuid.Nil, httpx.NewError(http.StatusForbidden, "plan_required", "buy a Flash plan to create a single-day flash event")
		}
		return uuid.Nil, httpx.NewError(http.StatusForbidden, "plan_required", "choose a plan to start creating events")
	}
	return uuid.Nil, httpx.NewError(http.StatusForbidden, "event_limit_reached", "you've reached your plan's event limit — upgrade to create more events")
}

func (s *Service) Create(ctx context.Context, orgID uuid.UUID, req CreateEventRequest) (EventResponse, error) {
	subscriptionID, err := s.checkEventLimit(ctx, orgID, req.ScheduleMode)
	if err != nil {
		return EventResponse{}, err
	}

	switch req.ScheduleMode {
	case "flash", "selective":
		return s.createExplicitDays(ctx, orgID, subscriptionID, req)
	case "fixed_range":
		return s.createFixedRange(ctx, orgID, subscriptionID, req)
	case "recurring":
		return s.createRecurring(ctx, orgID, subscriptionID, req)
	default:
		return EventResponse{}, httpx.ErrValidation(map[string]string{"schedule_mode": "unknown schedule mode"})
	}
}

func (s *Service) createExplicitDays(ctx context.Context, orgID, subscriptionID uuid.UUID, req CreateEventRequest) (EventResponse, error) {
	days, err := parseDays(req.Days)
	if err != nil {
		return EventResponse{}, err
	}
	if len(days) == 0 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"days": "at least one day is required"})
	}
	if req.ScheduleMode == "flash" && len(days) != 1 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
	}
	minDate, maxDate := minMaxDates(days)

	var event dbgen.Event
	var createdDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		created, err := repo.Create(ctx, orgID, subscriptionID, req.Name, req.ScheduleMode, minDate, maxDate, toPgText(req.Venue))
		if err != nil {
			return httpx.ErrInternal()
		}
		event = created
		createdDays, err = writeDaysTx(ctx, repo, event.ID, days)
		return err
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return toEventResponse(event, createdDays, nil, nil), nil
}

func (s *Service) createFixedRange(ctx context.Context, orgID, subscriptionID uuid.UUID, req CreateEventRequest) (EventResponse, error) {
	days, excluded, err := expandFixedRange(req.RangeStart, req.RangeEnd, req.RangeEntryTime, req.RangeExitTime, req.ExcludedDates)
	if err != nil {
		return EventResponse{}, err
	}
	minDate, maxDate := minMaxDates(days)

	var event dbgen.Event
	var createdDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		created, err := repo.Create(ctx, orgID, subscriptionID, req.Name, "fixed_range", minDate, maxDate, toPgText(req.Venue))
		if err != nil {
			return httpx.ErrInternal()
		}
		event = created
		createdDays, err = writeDaysTx(ctx, repo, event.ID, days)
		if err != nil {
			return err
		}
		return writeExcludedDatesTx(ctx, repo, event.ID, excluded)
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return toEventResponse(event, createdDays, nil, formatDates(excluded)), nil
}

func (s *Service) createRecurring(ctx context.Context, orgID, subscriptionID uuid.UUID, req CreateEventRequest) (EventResponse, error) {
	rule, err := parseRecurrence(req.Recurrence, req.ExcludedDates)
	if err != nil {
		return EventResponse{}, err
	}
	horizonEnd := todayDate().AddDate(0, 0, recurringHorizonDays)
	days := materializeRecurringDays(rule, rule.startsOn.Time, horizonEnd)
	if len(days) == 0 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"recurrence": "no active dates in the next year — check weekdays and exclusions"})
	}

	var event dbgen.Event
	var createdDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		created, err := repo.Create(ctx, orgID, subscriptionID, req.Name, "recurring", rule.startsOn, rule.endsOn, toPgText(req.Venue))
		if err != nil {
			return httpx.ErrInternal()
		}
		event = created
		if _, err := repo.UpsertRecurrence(ctx, event.ID, rule.startsOn, rule.endsOn); err != nil {
			return httpx.ErrInternal()
		}
		if err := writeRecurrenceWeekdaysTx(ctx, repo, event.ID, rule.weekdays); err != nil {
			return err
		}
		if err := writeExcludedDatesTx(ctx, repo, event.ID, rule.excludedParsed); err != nil {
			return err
		}
		createdDays, err = writeDaysTx(ctx, repo, event.ID, days)
		return err
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return toEventResponse(event, createdDays, rule.toResponse(), formatDates(rule.excludedParsed)), nil
}

func (s *Service) Get(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
	event, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return EventResponse{}, httpx.ErrNotFound("event")
	}
	eventDays, err := s.repo.ListDays(ctx, id)
	if err != nil {
		return EventResponse{}, httpx.ErrInternal()
	}

	var recResp *RecurrenceResponse
	var excludedStrs []string
	switch event.ScheduleMode {
	case "recurring":
		recResp, err = s.loadRecurrenceResponse(ctx, id)
		if err != nil {
			return EventResponse{}, err
		}
		excludedStrs, err = s.loadExcludedDates(ctx, id)
		if err != nil {
			return EventResponse{}, err
		}
	case "fixed_range":
		excludedStrs, err = s.loadExcludedDates(ctx, id)
		if err != nil {
			return EventResponse{}, err
		}
	}
	return toEventResponse(event, eventDays, recResp, excludedStrs), nil
}

func (s *Service) loadRecurrenceResponse(ctx context.Context, eventID uuid.UUID) (*RecurrenceResponse, error) {
	rec, err := s.repo.GetRecurrence(ctx, eventID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	weekdays, err := s.repo.ListRecurrenceWeekdays(ctx, eventID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	resp := &RecurrenceResponse{
		StartsOn: timeutil.FormatDate(rec.StartsOn),
		Weekdays: make([]RecurrenceWeekdayResponse, len(weekdays)),
	}
	if rec.EndsOn.Valid {
		s := timeutil.FormatDate(rec.EndsOn)
		resp.EndsOn = &s
	}
	for i, w := range weekdays {
		resp.Weekdays[i] = RecurrenceWeekdayResponse{
			Weekday: int(w.Weekday), EntryTime: timeutil.FormatClock(w.EntryTime), ExitTime: timeutil.FormatClock(w.ExitTime),
		}
	}
	return resp, nil
}

func (s *Service) loadExcludedDates(ctx context.Context, eventID uuid.UUID) ([]string, error) {
	rows, err := s.repo.ListExcludedDates(ctx, eventID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = timeutil.FormatDate(r.Date)
	}
	return out, nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]EventSummary, error) {
	rows, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	summaries := make([]EventSummary, len(rows))
	for i, row := range rows {
		summaries[i] = toEventSummary(row)
	}
	return summaries, nil
}

// Update replaces a draft event's schedule — which fields are read
// depends on the event's existing schedule_mode, which can't itself
// change after creation (same invariant as kind before it).
func (s *Service) Update(ctx context.Context, orgID, id uuid.UUID, req UpdateEventRequest) (EventResponse, error) {
	existing, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return EventResponse{}, httpx.ErrNotFound("event")
	}
	if existing.Status != "draft" {
		return EventResponse{}, httpx.ErrConflict("only a draft event can be edited")
	}

	switch existing.ScheduleMode {
	case "flash", "selective":
		return s.updateExplicitDays(ctx, orgID, id, existing, req)
	case "fixed_range":
		return s.updateFixedRange(ctx, orgID, id, req)
	case "recurring":
		return s.updateRecurring(ctx, orgID, id, req)
	default:
		return EventResponse{}, httpx.ErrInternal()
	}
}

func (s *Service) updateExplicitDays(ctx context.Context, orgID, id uuid.UUID, existing dbgen.Event, req UpdateEventRequest) (EventResponse, error) {
	days, err := parseDays(req.Days)
	if err != nil {
		return EventResponse{}, err
	}
	if len(days) == 0 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"days": "at least one day is required"})
	}
	if existing.ScheduleMode == "flash" && len(days) != 1 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
	}
	minDate, maxDate := minMaxDates(days)

	var event dbgen.Event
	var updatedDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		updated, err := repo.UpdateNameAndDates(ctx, orgID, id, req.Name, minDate, maxDate, toPgText(req.Venue))
		if err != nil {
			// Not found, wrong org, or raced with a publish between the
			// pre-check above and here — all collapse to the same 409.
			return httpx.ErrConflict("only a draft event can be edited")
		}
		event = updated
		updatedDays, err = writeDaysTx(ctx, repo, id, days)
		return err
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return toEventResponse(event, updatedDays, nil, nil), nil
}

func (s *Service) updateFixedRange(ctx context.Context, orgID, id uuid.UUID, req UpdateEventRequest) (EventResponse, error) {
	days, excluded, err := expandFixedRange(req.RangeStart, req.RangeEnd, req.RangeEntryTime, req.RangeExitTime, req.ExcludedDates)
	if err != nil {
		return EventResponse{}, err
	}
	minDate, maxDate := minMaxDates(days)

	var event dbgen.Event
	var updatedDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		updated, err := repo.UpdateNameAndDates(ctx, orgID, id, req.Name, minDate, maxDate, toPgText(req.Venue))
		if err != nil {
			return httpx.ErrConflict("only a draft event can be edited")
		}
		event = updated
		updatedDays, err = writeDaysTx(ctx, repo, id, days)
		if err != nil {
			return err
		}
		return writeExcludedDatesTx(ctx, repo, id, excluded)
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return toEventResponse(event, updatedDays, nil, formatDates(excluded)), nil
}

func (s *Service) updateRecurring(ctx context.Context, orgID, id uuid.UUID, req UpdateEventRequest) (EventResponse, error) {
	rule, err := parseRecurrence(req.Recurrence, req.ExcludedDates)
	if err != nil {
		return EventResponse{}, err
	}
	horizonEnd := todayDate().AddDate(0, 0, recurringHorizonDays)
	days := materializeRecurringDays(rule, rule.startsOn.Time, horizonEnd)
	if len(days) == 0 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"recurrence": "no active dates in the next year — check weekdays and exclusions"})
	}

	var event dbgen.Event
	var updatedDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		updated, err := repo.UpdateNameAndDates(ctx, orgID, id, req.Name, rule.startsOn, rule.endsOn, toPgText(req.Venue))
		if err != nil {
			return httpx.ErrConflict("only a draft event can be edited")
		}
		event = updated
		if _, err := repo.UpsertRecurrence(ctx, id, rule.startsOn, rule.endsOn); err != nil {
			return httpx.ErrInternal()
		}
		if err := writeRecurrenceWeekdaysTx(ctx, repo, id, rule.weekdays); err != nil {
			return err
		}
		if err := writeExcludedDatesTx(ctx, repo, id, rule.excludedParsed); err != nil {
			return err
		}
		updatedDays, err = writeDaysTx(ctx, repo, id, days)
		return err
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return toEventResponse(event, updatedDays, rule.toResponse(), formatDates(rule.excludedParsed)), nil
}

// Publish locks the event's schedule and backfills joined_at (see
// people.joined_at, and 000016_people_joined_at_nullable) for everyone
// added while it was still a draft. For a recurring event, publish
// freezes the weekday/time rule and start/end date exactly like the
// other modes — but unlike them, the background horizon-extension job
// (see ExtendRecurringHorizon) keeps materializing new event_days for it
// afterward, since that's a system process rather than an admin edit, and
// AddExcludedDate remains available post-publish as the one narrow
// exception to "only a draft event can be edited".
func (s *Service) Publish(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		published, err := repo.Publish(ctx, orgID, id)
		if err != nil {
			return httpx.NewError(http.StatusConflict, "not_publishable", "event not found or already published")
		}
		return repo.MarkUnjoinedPeopleJoinedAt(ctx, id, published.PublishedAt)
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return s.Get(ctx, orgID, id)
}

// AddExcludedDate adds a single unplanned holiday to a recurring event —
// the one schedule edit allowed regardless of draft/published status (see
// Publish's doc comment). It only ever adds; removing an exclusion
// requires the event still being a draft, via the normal Update path.
func (s *Service) AddExcludedDate(ctx context.Context, orgID, id uuid.UUID, req AddExcludedDateRequest) (EventResponse, error) {
	existing, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return EventResponse{}, httpx.ErrNotFound("event")
	}
	if existing.ScheduleMode != "recurring" {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"schedule_mode": "exclusion dates can only be added to a recurring event"})
	}
	date, err := timeutil.ParseDate(req.Date)
	if err != nil {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"date": "invalid date"})
	}

	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		if _, err := repo.CreateExcludedDate(ctx, id, date); err != nil {
			return httpx.ErrInternal()
		}
		return repo.DeleteDay(ctx, id, date)
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return s.Get(ctx, orgID, id)
}

func (s *Service) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	deleted, err := s.repo.DeleteDraft(ctx, orgID, id)
	if err != nil {
		return httpx.ErrInternal()
	}
	if !deleted {
		return httpx.NewError(http.StatusConflict, "not_deletable", "event not found or no longer a draft")
	}
	return nil
}

// DeleteForSubscription hard-deletes every event subscriptionID funded —
// called by internal/jobs once that subscription's grace period (or, for
// Flash, its post-event retention window) has expired. Unlike Delete,
// this isn't gated on draft status or org-scoped: a billing sweep is a
// system operation, not a user request. Plain error, not httpx — this
// method is never handler-facing.
func (s *Service) DeleteForSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	return s.repo.DeleteForSubscription(ctx, subscriptionID)
}

// EndDateForSubscription returns the end_date of the event a Flash
// subscription funds (each funds exactly one), or ok=false if no event
// has been created under it yet. A flash event's end_date is never NULL
// (schedule_mode='flash' forces start_date=end_date), so the nullable
// end_date introduced for recurring events doesn't affect this. Plain
// error, not httpx — used only by internal/jobs's Flash-retention sweep,
// never handler-facing.
func (s *Service) EndDateForSubscription(ctx context.Context, subscriptionID uuid.UUID) (t time.Time, ok bool, err error) {
	date, ok, err := s.repo.EndDateForSubscription(ctx, subscriptionID)
	if err != nil || !ok {
		return time.Time{}, ok, err
	}
	return date.Time, true, nil
}

// ExtendRecurringHorizon advances every published recurring event's
// materialized event_days window out to recurringHorizonDays from today —
// see internal/jobs, which calls this once a day. Each event resumes
// from its own latest materialized day (or its recurrence's starts_on, if
// none yet), so a delayed run just catches up further next time rather
// than losing coverage.
func (s *Service) ExtendRecurringHorizon(ctx context.Context) error {
	today := todayDate()
	events, err := s.repo.ListRecurringForMaterialization(ctx, pgtype.Date{Time: today, Valid: true})
	if err != nil {
		return err
	}
	horizonEnd := today.AddDate(0, 0, recurringHorizonDays)
	for _, event := range events {
		if err := s.extendOneRecurringEvent(ctx, event, horizonEnd); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) extendOneRecurringEvent(ctx context.Context, event dbgen.Event, horizonEnd time.Time) error {
	rec, err := s.repo.GetRecurrence(ctx, event.ID)
	if err != nil {
		return err
	}
	weekdayRows, err := s.repo.ListRecurrenceWeekdays(ctx, event.ID)
	if err != nil {
		return err
	}
	excludedRows, err := s.repo.ListExcludedDates(ctx, event.ID)
	if err != nil {
		return err
	}

	rule := recurrenceRule{
		startsOn: rec.StartsOn,
		endsOn:   rec.EndsOn,
		weekdays: make(map[int]timeWindow, len(weekdayRows)),
		excluded: make(map[string]bool, len(excludedRows)),
	}
	for _, w := range weekdayRows {
		rule.weekdays[int(w.Weekday)] = timeWindow{entry: w.EntryTime, exit: w.ExitTime}
	}
	for _, e := range excludedRows {
		rule.excluded[timeutil.FormatDate(e.Date)] = true
	}

	from := rule.startsOn.Time
	latest, ok, err := s.repo.LatestDay(ctx, event.ID)
	if err != nil {
		return err
	}
	if ok {
		from = latest.Time.AddDate(0, 0, 1)
	}
	if from.After(horizonEnd) {
		return nil // already caught up
	}

	for _, d := range materializeRecurringDays(rule, from, horizonEnd) {
		if _, err := s.repo.CreateDay(ctx, event.ID, d.date, d.entryTime, d.exitTime); err != nil {
			return err
		}
	}
	return nil
}

// EventContext is the minimal parent-event read shared by every module
// nested under an event (subevents, people, forms): is it still editable,
// and which dates can a sub-event/day-scoped child legally use. Unchanged
// by the schedule-mode rewrite — event_days stays the one materialized
// source of truth regardless of how an event authors its schedule.
type EventContext struct {
	Status string
	Dates  map[string]bool
}

func (c EventContext) IsDraft() bool { return c.Status == "draft" }

func (s *Service) GetContext(ctx context.Context, orgID, eventID uuid.UUID) (EventContext, error) {
	event, err := s.repo.Get(ctx, orgID, eventID)
	if err != nil {
		return EventContext{}, httpx.ErrNotFound("event")
	}
	eventDays, err := s.repo.ListDays(ctx, eventID)
	if err != nil {
		return EventContext{}, httpx.ErrInternal()
	}

	dates := make(map[string]bool, len(eventDays))
	for _, d := range eventDays {
		dates[timeutil.FormatDate(d.Date)] = true
	}
	return EventContext{Status: event.Status, Dates: dates}, nil
}

var dayCSVColumns = []string{"date", "entry_time", "exit_time"}

// ImportDaysCSV wholesale-replaces a draft fixed_range/selective event's
// day list from a CSV upload — the same replace semantics as Update
// (not an upsert), just fed from a file instead of a form. Each row is
// validated independently and a bad one is skipped and reported rather
// than failing the whole file (mirrors people.Service.ImportCSV). The
// uploaded file itself is never persisted.
func (s *Service) ImportDaysCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader) (DayImportSummary, error) {
	existing, err := s.repo.Get(ctx, orgID, eventID)
	if err != nil {
		return DayImportSummary{}, httpx.ErrNotFound("event")
	}
	if existing.Status != "draft" {
		return DayImportSummary{}, httpx.ErrConflict("only a draft event can be edited")
	}
	if existing.ScheduleMode != "fixed_range" && existing.ScheduleMode != "selective" {
		return DayImportSummary{}, httpx.ErrValidation(map[string]string{"schedule_mode": "day import is only available for fixed-range or selective events"})
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return DayImportSummary{}, httpx.NewError(http.StatusBadRequest, "bad_csv", "could not read CSV header")
	}
	columns := make(map[string]int, len(header))
	for i, col := range header {
		columns[strings.ToLower(strings.TrimSpace(col))] = i
	}
	for _, required := range dayCSVColumns {
		if _, ok := columns[required]; !ok {
			return DayImportSummary{}, httpx.NewError(http.StatusBadRequest, "bad_csv", "missing required column: "+required)
		}
	}

	summary := DayImportSummary{}
	seen := make(map[string]bool)
	var days []parsedDay
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "malformed row"})
			continue
		}

		field := func(col string) string {
			idx, ok := columns[col]
			if !ok || idx >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[idx])
		}

		dateStr, entryStr, exitStr := field("date"), field("entry_time"), field("exit_time")
		if dateStr == "" || entryStr == "" || exitStr == "" {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "missing date, entry_time, or exit_time"})
			continue
		}
		if seen[dateStr] {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "duplicate date: " + dateStr})
			continue
		}
		date, err := timeutil.ParseDate(dateStr)
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "invalid date"})
			continue
		}
		entry, exit, err := parseTimeWindow(entryStr, exitStr)
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "invalid entry_time/exit_time"})
			continue
		}
		seen[dateStr] = true
		days = append(days, parsedDay{date: date, entryTime: entry, exitTime: exit})
	}

	if existing.ScheduleMode == "flash" && len(days) != 1 { // unreachable today (flash isn't importable) but keeps this in lockstep if that ever changes
		return DayImportSummary{}, httpx.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
	}
	if len(days) == 0 {
		return DayImportSummary{}, httpx.ErrValidation(map[string]string{"days": "no valid rows to import"})
	}
	minDate, maxDate := minMaxDates(days)

	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		if _, err := repo.UpdateNameAndDates(ctx, orgID, eventID, existing.Name, minDate, maxDate, existing.Venue); err != nil {
			return httpx.ErrConflict("only a draft event can be edited")
		}
		_, err := writeDaysTx(ctx, repo, eventID, days)
		return err
	})
	if txErr != nil {
		return DayImportSummary{}, txErr
	}

	summary.Imported = len(days)
	return summary, nil
}

// ImportExcludedDatesCSV adds every valid date in the upload as an
// exclusion (an upsert per row, like AddExcludedDate — never a wholesale
// replace) — available for fixed_range (draft only) and recurring (any
// status, same post-publish carve-out as AddExcludedDate).
func (s *Service) ImportExcludedDatesCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader) (DayImportSummary, error) {
	existing, err := s.repo.Get(ctx, orgID, eventID)
	if err != nil {
		return DayImportSummary{}, httpx.ErrNotFound("event")
	}
	if existing.ScheduleMode != "fixed_range" && existing.ScheduleMode != "recurring" {
		return DayImportSummary{}, httpx.ErrValidation(map[string]string{"schedule_mode": "exclusion import is only available for fixed-range or recurring events"})
	}
	if existing.ScheduleMode == "fixed_range" && existing.Status != "draft" {
		return DayImportSummary{}, httpx.ErrConflict("only a draft event can be edited")
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return DayImportSummary{}, httpx.NewError(http.StatusBadRequest, "bad_csv", "could not read CSV header")
	}
	columns := make(map[string]int, len(header))
	for i, col := range header {
		columns[strings.ToLower(strings.TrimSpace(col))] = i
	}
	dateIdx, ok := columns["date"]
	if !ok {
		return DayImportSummary{}, httpx.NewError(http.StatusBadRequest, "bad_csv", "missing required column: date")
	}

	summary := DayImportSummary{}
	var dates []pgtype.Date
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "malformed row"})
			continue
		}
		raw := ""
		if dateIdx < len(record) {
			raw = strings.TrimSpace(record[dateIdx])
		}
		if raw == "" {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "missing date"})
			continue
		}
		date, err := timeutil.ParseDate(raw)
		if err != nil {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "invalid date"})
			continue
		}
		dates = append(dates, date)
	}

	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)
		for _, d := range dates {
			if _, err := repo.CreateExcludedDate(ctx, eventID, d); err != nil {
				return httpx.ErrInternal()
			}
			if err := repo.DeleteDay(ctx, eventID, d); err != nil {
				return httpx.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return DayImportSummary{}, txErr
	}

	summary.Imported = len(dates)
	return summary, nil
}

// ExportDays returns a CSV file, not the usual JSON envelope — mirrors
// people.Service.Export. Always available regardless of schedule_mode,
// since it reads straight from event_days, the one materialized truth
// every mode ultimately produces.
func (s *Service) ExportDays(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, error) {
	if _, err := s.repo.Get(ctx, orgID, eventID); err != nil {
		return nil, httpx.ErrNotFound("event")
	}
	days, err := s.repo.ListDays(ctx, eventID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"date", "entry_time", "exit_time"})
	for _, d := range days {
		_ = w.Write([]string{timeutil.FormatDate(d.Date), timeutil.FormatClock(d.EntryTime), timeutil.FormatClock(d.ExitTime)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, httpx.ErrInternal()
	}
	return buf.Bytes(), nil
}

// -- shared per-tx writers (Create and a draft-only Update funnel through
// the same replace-everything-for-this-event write path) --

func writeDaysTx(ctx context.Context, repo *Repository, eventID uuid.UUID, days []parsedDay) ([]dbgen.EventDay, error) {
	if err := repo.DeleteDaysForEvent(ctx, eventID); err != nil {
		return nil, httpx.ErrInternal()
	}
	out := make([]dbgen.EventDay, 0, len(days))
	for _, d := range days {
		day, err := repo.CreateDay(ctx, eventID, d.date, d.entryTime, d.exitTime)
		if err != nil {
			return nil, httpx.ErrInternal()
		}
		out = append(out, day)
	}
	return out, nil
}

func writeExcludedDatesTx(ctx context.Context, repo *Repository, eventID uuid.UUID, dates []pgtype.Date) error {
	if err := repo.DeleteExcludedDatesForEvent(ctx, eventID); err != nil {
		return httpx.ErrInternal()
	}
	for _, d := range dates {
		if _, err := repo.CreateExcludedDate(ctx, eventID, d); err != nil {
			return httpx.ErrInternal()
		}
	}
	return nil
}

func writeRecurrenceWeekdaysTx(ctx context.Context, repo *Repository, eventID uuid.UUID, weekdays map[int]timeWindow) error {
	if err := repo.DeleteRecurrenceWeekdaysForEvent(ctx, eventID); err != nil {
		return httpx.ErrInternal()
	}
	for wd, win := range weekdays {
		if _, err := repo.CreateRecurrenceWeekday(ctx, eventID, int16(wd), win.entry, win.exit); err != nil {
			return httpx.ErrInternal()
		}
	}
	return nil
}

// -- explicit-day parsing (flash/selective) --

type parsedDay struct {
	date      pgtype.Date
	entryTime pgtype.Time
	exitTime  pgtype.Time
}

func parseDays(inputs []EventDayInput) ([]parsedDay, error) {
	seen := make(map[string]bool, len(inputs))
	days := make([]parsedDay, len(inputs))

	for i, in := range inputs {
		if seen[in.Date] {
			return nil, httpx.ErrValidation(map[string]string{"days": "duplicate date: " + in.Date})
		}
		seen[in.Date] = true

		date, err := timeutil.ParseDate(in.Date)
		if err != nil {
			return nil, httpx.ErrValidation(map[string]string{"days": "invalid date: " + in.Date})
		}
		entry, exit, err := parseTimeWindow(in.EntryTime, in.ExitTime)
		if err != nil {
			return nil, httpx.ErrValidation(map[string]string{"days": "invalid time on " + in.Date})
		}

		days[i] = parsedDay{date: date, entryTime: entry, exitTime: exit}
	}
	return days, nil
}

func minMaxDates(days []parsedDay) (pgtype.Date, pgtype.Date) {
	min, max := days[0].date, days[0].date
	for _, d := range days[1:] {
		if d.date.Time.Before(min.Time) {
			min = d.date
		}
		if d.date.Time.After(max.Time) {
			max = d.date
		}
	}
	return min, max
}

// -- fixed_range expansion --

func expandFixedRange(rangeStart, rangeEnd, entryTimeStr, exitTimeStr string, excludedDates []string) ([]parsedDay, []pgtype.Date, error) {
	start, err := timeutil.ParseDate(rangeStart)
	if err != nil {
		return nil, nil, httpx.ErrValidation(map[string]string{"range_start": "invalid date"})
	}
	if start.Time.Before(todayDate()) {
		return nil, nil, httpx.ErrValidation(map[string]string{"range_start": "must be today or in the future"})
	}
	end, err := timeutil.ParseDate(rangeEnd)
	if err != nil {
		return nil, nil, httpx.ErrValidation(map[string]string{"range_end": "invalid date"})
	}
	if end.Time.Before(start.Time) {
		return nil, nil, httpx.ErrValidation(map[string]string{"range_end": "must be on or after range_start"})
	}
	if int(end.Time.Sub(start.Time).Hours()/24) > fixedRangeMaxSpanDays {
		return nil, nil, httpx.ErrValidation(map[string]string{"range_end": "a fixed date range can span at most 2 years"})
	}
	entry, exit, err := parseTimeWindow(entryTimeStr, exitTimeStr)
	if err != nil {
		return nil, nil, err
	}

	excluded, excludedParsed, err := parseExcludedDates(excludedDates)
	if err != nil {
		return nil, nil, err
	}

	var days []parsedDay
	for cur := start.Time; !cur.After(end.Time); cur = cur.AddDate(0, 0, 1) {
		if excluded[cur.Format(timeutil.DateLayout)] {
			continue
		}
		days = append(days, parsedDay{date: pgtype.Date{Time: cur, Valid: true}, entryTime: entry, exitTime: exit})
	}
	if len(days) == 0 {
		return nil, nil, httpx.ErrValidation(map[string]string{"excluded_dates": "excludes every date in the range"})
	}
	return days, excludedParsed, nil
}

// -- recurring rule parsing/materialization --

type timeWindow struct {
	entry, exit pgtype.Time
}

type recurrenceRule struct {
	startsOn       pgtype.Date
	endsOn         pgtype.Date // Valid=false => open-ended
	weekdays       map[int]timeWindow
	excluded       map[string]bool
	excludedParsed []pgtype.Date
}

func (r recurrenceRule) toResponse() *RecurrenceResponse {
	resp := &RecurrenceResponse{StartsOn: timeutil.FormatDate(r.startsOn)}
	if r.endsOn.Valid {
		s := timeutil.FormatDate(r.endsOn)
		resp.EndsOn = &s
	}
	weekdays := make([]int, 0, len(r.weekdays))
	for wd := range r.weekdays {
		weekdays = append(weekdays, wd)
	}
	sort.Ints(weekdays)
	resp.Weekdays = make([]RecurrenceWeekdayResponse, len(weekdays))
	for i, wd := range weekdays {
		win := r.weekdays[wd]
		resp.Weekdays[i] = RecurrenceWeekdayResponse{Weekday: wd, EntryTime: timeutil.FormatClock(win.entry), ExitTime: timeutil.FormatClock(win.exit)}
	}
	return resp
}

func parseRecurrence(in *RecurrenceInput, excludedDates []string) (recurrenceRule, error) {
	if in == nil {
		return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence": "required for a recurring event"})
	}
	startsOn, err := timeutil.ParseDate(in.StartsOn)
	if err != nil {
		return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence.starts_on": "invalid date"})
	}
	if startsOn.Time.Before(todayDate()) {
		return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence.starts_on": "must be today or in the future"})
	}

	var endsOn pgtype.Date
	if in.EndsOn != nil {
		parsed, err := timeutil.ParseDate(*in.EndsOn)
		if err != nil {
			return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence.ends_on": "invalid date"})
		}
		if parsed.Time.Before(startsOn.Time) {
			return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence.ends_on": "must be on or after starts_on"})
		}
		endsOn = parsed
	}

	weekdays := make(map[int]timeWindow, len(in.Weekdays))
	for _, w := range in.Weekdays {
		if w.Weekday < 0 || w.Weekday > 6 {
			return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence.weekdays": "weekday must be 0-6"})
		}
		if _, dup := weekdays[w.Weekday]; dup {
			return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence.weekdays": "duplicate weekday"})
		}
		entry, exit, err := parseTimeWindow(w.EntryTime, w.ExitTime)
		if err != nil {
			return recurrenceRule{}, err
		}
		weekdays[w.Weekday] = timeWindow{entry: entry, exit: exit}
	}
	if len(weekdays) == 0 {
		return recurrenceRule{}, httpx.ErrValidation(map[string]string{"recurrence.weekdays": "at least one active weekday is required"})
	}

	excluded, excludedParsed, err := parseExcludedDates(excludedDates)
	if err != nil {
		return recurrenceRule{}, err
	}

	return recurrenceRule{startsOn: startsOn, endsOn: endsOn, weekdays: weekdays, excluded: excluded, excludedParsed: excludedParsed}, nil
}

func parseExcludedDates(raw []string) (map[string]bool, []pgtype.Date, error) {
	excluded := make(map[string]bool, len(raw))
	parsed := make([]pgtype.Date, 0, len(raw))
	for _, s := range raw {
		date, err := timeutil.ParseDate(s)
		if err != nil {
			return nil, nil, httpx.ErrValidation(map[string]string{"excluded_dates": "invalid date: " + s})
		}
		if !excluded[s] {
			excluded[s] = true
			parsed = append(parsed, date)
		}
	}
	return excluded, parsed, nil
}

func parseTimeWindow(entryStr, exitStr string) (pgtype.Time, pgtype.Time, error) {
	entry, err := timeutil.ParseClock(entryStr)
	if err != nil {
		return pgtype.Time{}, pgtype.Time{}, httpx.ErrValidation(map[string]string{"entry_time": "invalid entry_time: " + entryStr})
	}
	exit, err := timeutil.ParseClock(exitStr)
	if err != nil {
		return pgtype.Time{}, pgtype.Time{}, httpx.ErrValidation(map[string]string{"exit_time": "invalid exit_time: " + exitStr})
	}
	if exit.Microseconds <= entry.Microseconds {
		return pgtype.Time{}, pgtype.Time{}, httpx.ErrValidation(map[string]string{"exit_time": "exit_time must be after entry_time"})
	}
	return entry, exit, nil
}

// materializeRecurringDays walks every date from from to horizonEnd
// (inclusive), stopping early at the rule's own ends_on if it has one,
// and emits a day for each date whose weekday is active and isn't
// excluded.
func materializeRecurringDays(rule recurrenceRule, from, horizonEnd time.Time) []parsedDay {
	from = dateOnly(from)
	horizonEnd = dateOnly(horizonEnd)
	var days []parsedDay
	for cur := from; !cur.After(horizonEnd); cur = cur.AddDate(0, 0, 1) {
		if rule.endsOn.Valid && cur.After(rule.endsOn.Time) {
			break
		}
		window, ok := rule.weekdays[mondayFirstWeekday(cur)]
		if !ok {
			continue
		}
		if rule.excluded[cur.Format(timeutil.DateLayout)] {
			continue
		}
		days = append(days, parsedDay{date: pgtype.Date{Time: cur, Valid: true}, entryTime: window.entry, exitTime: window.exit})
	}
	return days
}

// mondayFirstWeekday converts Go's time.Weekday (Sunday=0..Saturday=6) to
// this package's wire convention (Monday=0..Sunday=6, see
// RecurrenceWeekdayInput).
func mondayFirstWeekday(t time.Time) int {
	return (int(t.Weekday()) + 6) % 7
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func todayDate() time.Time {
	return dateOnly(time.Now().UTC())
}

func formatDates(dates []pgtype.Date) []string {
	if len(dates) == 0 {
		return nil
	}
	out := make([]string, len(dates))
	for i, d := range dates {
		out[i] = timeutil.FormatDate(d)
	}
	return out
}

// -- response/wire conversion --

func toEventSummary(event dbgen.Event) EventSummary {
	var publishedAt *string
	if event.PublishedAt.Valid {
		s := event.PublishedAt.Time.Format(time.RFC3339)
		publishedAt = &s
	}
	var venue *string
	if event.Venue.Valid {
		venue = &event.Venue.String
	}
	var endDate *string
	if event.EndDate.Valid {
		s := timeutil.FormatDate(event.EndDate)
		endDate = &s
	}
	return EventSummary{
		ID:           event.ID,
		Name:         event.Name,
		ScheduleMode: event.ScheduleMode,
		Status:       event.Status,
		StartDate:    timeutil.FormatDate(event.StartDate),
		EndDate:      endDate,
		Venue:        venue,
		PublishedAt:  publishedAt,
	}
}

func toPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func toEventResponse(event dbgen.Event, eventDays []dbgen.EventDay, recurrence *RecurrenceResponse, excludedDates []string) EventResponse {
	dayResponses := make([]EventDayResponse, len(eventDays))
	for i, d := range eventDays {
		dayResponses[i] = EventDayResponse{
			Date:      timeutil.FormatDate(d.Date),
			EntryTime: timeutil.FormatClock(d.EntryTime),
			ExitTime:  timeutil.FormatClock(d.ExitTime),
		}
	}
	return EventResponse{
		EventSummary:  toEventSummary(event),
		Days:          dayResponses,
		Recurrence:    recurrence,
		ExcludedDates: excludedDates,
	}
}
