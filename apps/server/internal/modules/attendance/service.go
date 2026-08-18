package attendance

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"identitycard-server/internal/config"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/events"
	"identitycard-server/internal/modules/people"
	"identitycard-server/internal/modules/subevents"
	"identitycard-server/internal/qrtoken"
	"identitycard-server/internal/timeutil"
)

// statusGrace is how close to the scheduled time counts as "on_time"
// rather than early/late — a scan isn't going to land on the exact minute.
const statusGrace = 10 * time.Minute

type Service struct {
	cfg       *config.Config
	repo      *Repository
	events    *events.Service
	people    *people.Service
	subevents *subevents.Service
}

func NewService(cfg *config.Config, repo *Repository, eventsService *events.Service, peopleService *people.Service, subEventsService *subevents.Service) *Service {
	return &Service{cfg: cfg, repo: repo, events: eventsService, people: peopleService, subevents: subEventsService}
}

// Scan is the whole point of the QR code: verify it's genuine and belongs
// to the scanning device's own organization, confirm the person is
// permitted in today (per the event's or their sub-events' schedule), and
// record an entry (first scan of the day) or exit (second) — a third scan
// the same day is read-only, returning the existing record rather than
// erroring, since a device operator re-scanning by mistake shouldn't be
// treated as a failure.
func (s *Service) Scan(ctx context.Context, deviceOrgID, deviceID uuid.UUID, qrToken string) (ScanResponse, error) {
	claims, err := qrtoken.Parse(s.cfg.QRSecret, qrToken)
	if err != nil {
		return ScanResponse{}, httpx.NewError(http.StatusBadRequest, "invalid_qr", "QR code is invalid or expired")
	}
	if claims.OrganizationID != deviceOrgID {
		return ScanResponse{}, httpx.ErrForbidden("this card belongs to a different organization")
	}

	event, err := s.events.Get(ctx, deviceOrgID, claims.EventID)
	if err != nil {
		return ScanResponse{}, httpx.ErrNotFound("event")
	}
	if event.Status != "published" {
		return ScanResponse{}, httpx.NewError(http.StatusConflict, "not_published", "event is not published yet")
	}

	person, err := s.people.Get(ctx, deviceOrgID, claims.EventID, claims.PersonID)
	if err != nil {
		return ScanResponse{}, httpx.ErrNotFound("person")
	}

	windows, err := s.resolveExpectedWindows(ctx, deviceOrgID, claims.EventID, person, event)
	if err != nil {
		return ScanResponse{}, err
	}
	today := time.Now().Format(timeutil.DateLayout)
	window, permitted := windows[today]
	if !permitted {
		return ScanResponse{}, httpx.NewError(http.StatusForbidden, "not_permitted_today", "this person is not permitted to enter today")
	}

	now := time.Now()

	existing, err := s.repo.Get(ctx, person.ID, today)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		status := classify(now, today, window.EntryTime)
		if _, err := s.repo.CreateEntry(ctx, deviceOrgID, claims.EventID, person.ID, today, now, status, deviceID); err != nil {
			return ScanResponse{}, httpx.ErrInternal()
		}
		return ScanResponse{Direction: "entry", Status: &status, Date: today, EventName: event.Name, Person: person}, nil

	case err != nil:
		return ScanResponse{}, httpx.ErrInternal()

	case !existing.ExitAt.Valid:
		status := classify(now, today, window.ExitTime)
		if _, err := s.repo.RecordExit(ctx, person.ID, today, now, status, deviceID); err != nil {
			return ScanResponse{}, httpx.ErrInternal()
		}
		return ScanResponse{Direction: "exit", Status: &status, Date: today, EventName: event.Name, Person: person}, nil

	default:
		return ScanResponse{Direction: "already_completed", Date: today, EventName: event.Name, Person: person}, nil
	}
}

// RosterFilter narrows BuildRoster's output. All fields are optional.
type RosterFilter struct {
	SubEventID *uuid.UUID
	Date       *string
	Attended   *bool
	Status     *string // matches either entry_status or exit_status
}

// BuildRoster is the one place that turns raw attendance_records into the
// admin-facing view — and, unlike a raw records list, it includes people
// who were never scanned at all (Attended: false, every *At/*Status nil),
// since "who's absent" is exactly what the spec's analytics ask for.
// analytics.Service reuses this unfiltered to compute its rollups, so
// there's only one place this roster-building logic lives.
func (s *Service) BuildRoster(ctx context.Context, orgID, eventID uuid.UUID, filter RosterFilter) ([]RosterEntry, error) {
	event, err := s.events.Get(ctx, orgID, eventID)
	if err != nil {
		return nil, err
	}
	roster, err := s.people.List(ctx, orgID, eventID, people.ListFilter{SubEventID: filter.SubEventID})
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	rows, err := s.repo.ListForEvent(ctx, orgID, eventID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}

	type recordKey struct {
		person uuid.UUID
		date   string
	}
	records := make(map[recordKey]dbgen.ListAttendanceForEventRow, len(rows))
	for _, row := range rows {
		records[recordKey{row.PersonID, timeutil.FormatDate(row.Date)}] = row
	}

	entries := make([]RosterEntry, 0)
	for _, person := range roster {
		windows, err := s.resolveExpectedWindows(ctx, orgID, eventID, person, event)
		if err != nil {
			return nil, err
		}
		for _, day := range event.Days {
			if filter.Date != nil && day.Date != *filter.Date {
				continue
			}
			if _, permitted := windows[day.Date]; !permitted {
				continue
			}

			entry := RosterEntry{PersonID: person.ID, PersonName: person.Name, PersonEmail: person.Email, Date: day.Date}
			if row, ok := records[recordKey{person.ID, day.Date}]; ok {
				entry = mergeRecord(entry, row)
			}

			if filter.Attended != nil && entry.Attended != *filter.Attended {
				continue
			}
			if filter.Status != nil {
				matches := (entry.EntryStatus != nil && *entry.EntryStatus == *filter.Status) ||
					(entry.ExitStatus != nil && *entry.ExitStatus == *filter.Status)
				if !matches {
					continue
				}
			}

			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// Export renders BuildRoster's output as CSV — same "generated on demand,
// never stored" shape as people.Service.Export.
func (s *Service) Export(ctx context.Context, orgID, eventID uuid.UUID, filter RosterFilter) ([]byte, error) {
	entries, err := s.BuildRoster(ctx, orgID, eventID, filter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"person_name", "person_email", "date", "attended", "entry_at", "entry_status", "exit_at", "exit_status"})
	for _, e := range entries {
		_ = w.Write([]string{
			e.PersonName, e.PersonEmail, e.Date, formatBool(e.Attended),
			derefOrEmpty(e.EntryAt), derefOrEmpty(e.EntryStatus),
			derefOrEmpty(e.ExitAt), derefOrEmpty(e.ExitStatus),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, httpx.ErrInternal()
	}
	return buf.Bytes(), nil
}

type window struct {
	EntryTime string
	ExitTime  string
}

// resolveExpectedWindows returns every date on which person is permitted
// to enter, and the entry/exit clock time that governs it: the event's
// own schedule if they have no sub-events, otherwise the widest window
// (earliest entry, latest exit) across every sub-event they belong to
// that runs on a given date — a sub-event with no days of its own
// inherits the whole event's (see subevents' Phase 2/3 design). Computed
// once per person rather than once per (person, date) — this is called
// from a loop over an event's days in BuildRoster, and each sub-event
// lookup is its own DB read.
func (s *Service) resolveExpectedWindows(ctx context.Context, orgID, eventID uuid.UUID, person people.PersonResponse, event events.EventResponse) (map[string]window, error) {
	windows := make(map[string]window, len(event.Days))

	if len(person.SubEventIDs) == 0 {
		for _, d := range event.Days {
			windows[d.Date] = window{EntryTime: d.EntryTime, ExitTime: d.ExitTime}
		}
		return windows, nil
	}

	for _, subEventID := range person.SubEventIDs {
		se, err := s.subevents.Get(ctx, orgID, eventID, subEventID)
		if err != nil {
			continue // stale link outliving its sub-event — skip rather than fail
		}
		days := se.Days
		if len(days) == 0 {
			days = event.Days
		}
		for _, d := range days {
			existing, ok := windows[d.Date]
			if !ok {
				windows[d.Date] = window{EntryTime: d.EntryTime, ExitTime: d.ExitTime}
				continue
			}
			if d.EntryTime < existing.EntryTime {
				existing.EntryTime = d.EntryTime
			}
			if d.ExitTime > existing.ExitTime {
				existing.ExitTime = d.ExitTime
			}
			windows[d.Date] = existing
		}
	}
	return windows, nil
}

// classify compares a scan against a scheduled "HH:MM" clock time on a
// given date, within statusGrace of it counting as on_time.
func classify(scanAt time.Time, date, scheduledClock string) string {
	scheduled, err := time.ParseInLocation(timeutil.DateLayout+" "+timeutil.ClockLayout, date+" "+scheduledClock, scanAt.Location())
	if err != nil {
		return "on_time"
	}
	diff := scanAt.Sub(scheduled)
	switch {
	case diff < -statusGrace:
		return "early"
	case diff > statusGrace:
		return "late"
	default:
		return "on_time"
	}
}

func mergeRecord(entry RosterEntry, row dbgen.ListAttendanceForEventRow) RosterEntry {
	entry.Attended = row.EntryAt.Valid
	if row.EntryAt.Valid {
		s := row.EntryAt.Time.Format(time.RFC3339)
		entry.EntryAt = &s
	}
	if row.EntryStatus.Valid {
		entry.EntryStatus = &row.EntryStatus.String
	}
	if row.ExitAt.Valid {
		s := row.ExitAt.Time.Format(time.RFC3339)
		entry.ExitAt = &s
	}
	if row.ExitStatus.Valid {
		entry.ExitStatus = &row.ExitStatus.String
	}
	return entry
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
