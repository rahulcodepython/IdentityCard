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

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/pkg/qrtoken"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

const statusGrace = 10 * time.Minute

func (a *App) Scan(ctx context.Context, deviceOrgID, deviceID uuid.UUID, qrToken string) (ScanResponse, error) {
	claims, err := qrtoken.Parse(a.cfg.QRSecret, qrToken)
	if err != nil {
		return ScanResponse{}, utils.NewError(http.StatusBadRequest, "invalid_qr", "QR code is invalid or expired")
	}
	if claims.OrganizationID != deviceOrgID {
		return ScanResponse{}, utils.ErrForbidden("this card belongs to a different organization")
	}

	event, err := a.events.Get(ctx, deviceOrgID, claims.EventID)
	if err != nil {
		return ScanResponse{}, utils.ErrNotFound("event")
	}
	if event.Status != "published" {
		return ScanResponse{}, utils.NewError(http.StatusConflict, "not_published", "event is not published yet")
	}

	person, err := a.people.Get(ctx, deviceOrgID, claims.EventID, claims.PersonID)
	if err != nil {
		return ScanResponse{}, utils.ErrNotFound("person")
	}

	windows, err := a.resolveExpectedWindows(ctx, deviceOrgID, claims.EventID, person, event)
	if err != nil {
		return ScanResponse{}, err
	}
	today := time.Now().Format(timeutil.DateLayout)
	window, permitted := windows[today]
	if !permitted {
		return ScanResponse{}, utils.NewError(http.StatusForbidden, "not_permitted_today", "this person is not permitted to enter today")
	}

	now := time.Now()

	existing, err := a.GetAttendanceRecord(ctx, person.ID, today)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		status := classify(now, today, window.EntryTime)
		if _, err := a.CreateAttendanceEntry(ctx, deviceOrgID, claims.EventID, person.ID, today, now, status, deviceID); err != nil {
			return ScanResponse{}, utils.ErrInternal()
		}
		return ScanResponse{Direction: "entry", Status: &status, Date: today, EventName: event.Name, Person: person}, nil

	case err != nil:
		return ScanResponse{}, utils.ErrInternal()

	case !existing.ExitAt.Valid:
		status := classify(now, today, window.ExitTime)
		if _, err := a.RecordAttendanceExit(ctx, person.ID, today, now, status, deviceID); err != nil {
			return ScanResponse{}, utils.ErrInternal()
		}
		return ScanResponse{Direction: "exit", Status: &status, Date: today, EventName: event.Name, Person: person}, nil

	default:
		return ScanResponse{Direction: "already_completed", Date: today, EventName: event.Name, Person: person}, nil
	}
}

func (a *App) BuildRoster(ctx context.Context, orgID, eventID uuid.UUID, filter RosterFilter) ([]RosterEntry, error) {
	event, err := a.events.Get(ctx, orgID, eventID)
	if err != nil {
		return nil, err
	}
	roster, err := a.people.List(ctx, orgID, eventID, people.PeopleListFilter{SubEventID: filter.SubEventID})
	if err != nil {
		return nil, utils.ErrInternal()
	}
	rows, err := a.ListAttendanceForEvent(ctx, orgID, eventID)
	if err != nil {
		return nil, utils.ErrInternal()
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
		windows, err := a.resolveExpectedWindows(ctx, orgID, eventID, person, event)
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

func (a *App) Export(ctx context.Context, orgID, eventID uuid.UUID, filter RosterFilter) ([]byte, error) {
	entries, err := a.BuildRoster(ctx, orgID, eventID, filter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"person_name", "person_email", "date", "attended", "entry_at", "entry_status", "exit_at", "exit_status"})
	for _, e := range entries {
		_ = w.Write([]string{
			e.PersonName, e.PersonEmail, e.Date, attendanceFormatBool(e.Attended),
			derefOrEmpty(e.EntryAt), derefOrEmpty(e.EntryStatus),
			derefOrEmpty(e.ExitAt), derefOrEmpty(e.ExitStatus),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, utils.ErrInternal()
	}
	return buf.Bytes(), nil
}

type attendanceWindow struct {
	EntryTime string
	ExitTime  string
}

func (a *App) resolveExpectedWindows(ctx context.Context, orgID, eventID uuid.UUID, person people.PersonResponse, event events.EventResponse) (map[string]attendanceWindow, error) {
	windows := make(map[string]attendanceWindow, len(event.Days))

	if len(person.SubEventIDs) == 0 {
		for _, d := range event.Days {
			windows[d.Date] = attendanceWindow{EntryTime: d.EntryTime, ExitTime: d.ExitTime}
		}
		return windows, nil
	}

	for _, subEventID := range person.SubEventIDs {
		se, err := a.subevents.Get(ctx, orgID, eventID, subEventID)
		if err != nil {
			continue
		}
		existing, ok := windows[se.Date]
		if !ok {
			windows[se.Date] = attendanceWindow{EntryTime: se.EntryTime, ExitTime: se.ExitTime}
			continue
		}
		if se.EntryTime < existing.EntryTime {
			existing.EntryTime = se.EntryTime
		}
		if se.ExitTime > existing.ExitTime {
			existing.ExitTime = se.ExitTime
		}
		windows[se.Date] = existing
	}
	return windows, nil
}

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

func attendanceFormatBool(b bool) string {
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
