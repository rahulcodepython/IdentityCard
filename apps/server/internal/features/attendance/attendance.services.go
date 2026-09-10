package attendance

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/pkg/qrtoken"
	"identitycard-server/internal/utils"
)

const statusGrace = 10 * time.Minute

func (a *App) Scan(ctx context.Context, deviceOrgID, deviceID uuid.UUID, qrToken string) (ScanResponse, error) {
	claims, err := qrtoken.Parse(a.cfg.QRSecret, qrToken)
	if err != nil {
		return ScanResponse{}, utils.NewError(http.StatusBadRequest, "QR code is invalid or expired.", generic.ErrAttendanceInvalidQR)
	}
	if claims.OrganizationID != deviceOrgID {
		return ScanResponse{}, utils.ErrForbidden("This card belongs to a different organization.", postgres.ErrForbidden)
	}

	event, err := a.events.Get(ctx, deviceOrgID, claims.EventID)
	if err != nil {
		return ScanResponse{}, utils.ErrNotFound("Event not found.", err)
	}
	if event.Status != "published" {
		return ScanResponse{}, utils.NewError(http.StatusConflict, "Event is not published yet.", generic.ErrEventsNotDraft)
	}

	person, err := a.people.Get(ctx, deviceOrgID, claims.EventID, claims.PersonID)
	if err != nil {
		return ScanResponse{}, utils.ErrNotFound("Person not found.", err)
	}

	windows, err := a.resolveExpectedWindows(ctx, deviceOrgID, claims.EventID, person, event)
	if err != nil {
		return ScanResponse{}, err
	}
	today := time.Now().Format(utils.DateLayout)
	window, permitted := windows[today]
	if !permitted {
		return ScanResponse{}, utils.NewError(http.StatusForbidden, "This person is not permitted to enter today.", nil)
	}

	now := time.Now()

	existing, err := a.GetAttendanceRecord(ctx, person.ID, today)
	switch {
	case errors.Is(err, postgres.ErrNotFound):
		status := classify(now, today, window.EntryTime)
		if _, err := a.CreateAttendanceEntry(ctx, deviceOrgID, claims.EventID, person.ID, today, now, status, deviceID); err != nil {
			return ScanResponse{}, utils.ErrInternal("Failed to record attendance entry.", err)
		}
		return ScanResponse{Direction: "entry", Status: &status, Date: today, EventName: event.Name, Person: person}, nil

	case err != nil:
		return ScanResponse{}, utils.ErrInternal("Failed to fetch attendance record.", err)

	case existing.ExitAt == nil:
		status := classify(now, today, window.ExitTime)
		if _, err := a.RecordAttendanceExit(ctx, person.ID, today, now, status, deviceID); err != nil {
			return ScanResponse{}, utils.ErrInternal("Failed to record attendance exit.", err)
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
		return nil, utils.ErrInternal("Failed to list people for roster.", err)
	}
	rows, err := a.ListAttendanceForEvent(ctx, orgID, eventID)
	if err != nil {
		return nil, utils.ErrInternal("Failed to list attendance records.", err)
	}

	type recordKey struct {
		person uuid.UUID
		date   string
	}
	records := make(map[recordKey]AttendanceRowDB, len(rows))
	for _, row := range rows {
		records[recordKey{row.PersonID, row.Date}] = row
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
		return nil, utils.ErrInternal("Failed to export roster CSV.", err)
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
	scheduled, err := time.ParseInLocation(utils.DateLayout+" "+utils.ClockLayout, date+" "+scheduledClock, scanAt.Location())
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

func mergeRecord(entry RosterEntry, row AttendanceRowDB) RosterEntry {
	entry.Attended = row.EntryAt != nil
	entry.EntryAt = row.EntryAt
	entry.EntryStatus = row.EntryStatus
	entry.ExitAt = row.ExitAt
	entry.ExitStatus = row.ExitStatus
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
