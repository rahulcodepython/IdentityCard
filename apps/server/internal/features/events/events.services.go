package events

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

const maxEventImageBytes = 2 << 20

var allowedEventImageContentTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
}

func (a *App) Create(ctx context.Context, orgID uuid.UUID, req CreateEventRequest) (EventResponse, error) {
	var days []parsedDay
	var startDate, endDate pgtype.Date
	var err error

	switch req.EventType {
	case "flash":
		if len(req.Days) != 1 {
			return EventResponse{}, utils.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
		}
		days, err = parseEventDays(req.Days)
		if err != nil {
			return EventResponse{}, err
		}
		startDate = days[0].date
		endDate = days[0].date

	case "standard":
		if len(req.Days) == 0 {
			return EventResponse{}, utils.ErrValidation(map[string]string{"days": "at least one day is required for a standard event"})
		}
		days, err = parseEventDays(req.Days)
		if err != nil {
			return EventResponse{}, err
		}
		startDate, endDate = minMaxEventDates(days)

	case "grouped":
		if req.RangeStart == "" || req.RangeEnd == "" {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "range_start and range_end are required for a grouped event"})
		}
		startDate, err = timeutil.ParseDate(req.RangeStart)
		if err != nil {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "invalid date"})
		}
		endDate, err = timeutil.ParseDate(req.RangeEnd)
		if err != nil {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "invalid date"})
		}
		if endDate.Time.Before(startDate.Time) {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "range_end must be on or after range_start"})
		}

	default:
		return EventResponse{}, utils.ErrValidation(map[string]string{"event_type": "unknown event type"})
	}

	var event dbgen.Event
	var meta dbgen.EventMetadatum
	var createdDays []dbgen.EventDay

	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		credit, err := a.plans.FindCreditTx(ctx, tx, orgID, req.EventType)
		if err != nil {
			return err
		}

		event, err = a.CreateEvent(ctx, tx, orgID, req.EventType, startDate, endDate)
		if err != nil {
			return utils.ErrInternal()
		}

		if len(days) > 0 {
			createdDays, err = a.writeEventDaysTx(ctx, tx, event.ID, days)
			if err != nil {
				return err
			}
		}

		meta, err = a.CreateMetadata(ctx, tx, event.ID, req.Name, textOrNull(req.Venue), textOrNull(req.OrganizerName))
		if err != nil {
			return utils.ErrInternal()
		}

		if err := a.plans.LinkCreditTx(ctx, tx, credit.ID, event.ID); err != nil {
			return err
		}

		return a.plans.MaybeReplenishUnlimitedTx(ctx, tx, orgID, credit.Type)
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}

	return toEventResponse(event, meta, createdDays), nil
}

func (a *App) Get(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
	event, err := a.GetEvent(ctx, orgID, id)
	if err != nil {
		return EventResponse{}, utils.ErrNotFound("event")
	}
	meta, err := a.GetMetadata(ctx, id)
	if err != nil {
		return EventResponse{}, utils.ErrInternal()
	}
	eventDays, err := a.ListDays(ctx, id)
	if err != nil {
		return EventResponse{}, utils.ErrInternal()
	}
	return toEventResponse(event, meta, eventDays), nil
}

func (a *App) List(ctx context.Context, orgID uuid.UUID) ([]EventSummary, error) {
	rows, err := a.ListEvents(ctx, orgID)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	summaries := make([]EventSummary, len(rows))
	for i, row := range rows {
		meta, err := a.GetMetadata(ctx, row.ID)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		summaries[i] = toEventSummary(row, meta)
	}
	return summaries, nil
}

func (a *App) Update(ctx context.Context, orgID, id uuid.UUID, req UpdateEventRequest) (EventResponse, error) {
	existing, err := a.GetEvent(ctx, orgID, id)
	if err != nil {
		return EventResponse{}, utils.ErrNotFound("event")
	}
	if existing.Status != "draft" {
		return EventResponse{}, utils.ErrConflict("only a draft event can be edited")
	}

	restricted, err := a.plans.GetCreditRestriction(ctx, id)
	if err != nil {
		return EventResponse{}, err
	}
	if restricted {
		return EventResponse{}, utils.NewError(http.StatusForbidden, "credit_restricted", "this event's billing is past due — renew your plan to edit")
	}

	var days []parsedDay
	var startDate, endDate pgtype.Date

	switch existing.EventType {
	case "flash":
		if len(req.Days) != 1 {
			return EventResponse{}, utils.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
		}
		days, err = parseEventDays(req.Days)
		if err != nil {
			return EventResponse{}, err
		}
		startDate = days[0].date
		endDate = days[0].date

	case "standard":
		if len(req.Days) == 0 {
			return EventResponse{}, utils.ErrValidation(map[string]string{"days": "at least one day is required for a standard event"})
		}
		days, err = parseEventDays(req.Days)
		if err != nil {
			return EventResponse{}, err
		}
		startDate, endDate = minMaxEventDates(days)

	case "grouped":
		if req.RangeStart == "" || req.RangeEnd == "" {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "range_start and range_end are required for a grouped event"})
		}
		startDate, err = timeutil.ParseDate(req.RangeStart)
		if err != nil {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "invalid date"})
		}
		endDate, err = timeutil.ParseDate(req.RangeEnd)
		if err != nil {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "invalid date"})
		}
		if endDate.Time.Before(startDate.Time) {
			return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "range_end must be on or after range_start"})
		}
	}

	var meta dbgen.EventMetadatum
	var updatedDays []dbgen.EventDay

	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		if _, err := a.UpdateDates(ctx, tx, orgID, id, startDate, endDate); err != nil {
			return utils.ErrConflict("only a draft event can be edited")
		}

		meta, err = a.UpdateMetadata(ctx, tx, id, req.Name, textOrNull(req.Venue), textOrNull(req.OrganizerName))
		if err != nil {
			return utils.ErrInternal()
		}

		if len(days) > 0 {
			updatedDays, err = a.writeEventDaysTx(ctx, tx, id, days)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}

	return toEventResponse(existing, meta, updatedDays), nil
}

func (a *App) Publish(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		published, err := a.PublishEvent(ctx, tx, orgID, id)
		if err != nil {
			return utils.NewError(http.StatusConflict, "not_publishable", "event not found or already published")
		}
		return a.MarkUnjoinedPeopleJoinedAt(ctx, tx, id, published.PublishedAt)
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}
	return a.Get(ctx, orgID, id)
}

func (a *App) DeleteDraftEvent(ctx context.Context, orgID, id uuid.UUID) error {
	deleted, err := a.DeleteDraft(ctx, orgID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !deleted {
		return utils.NewError(http.StatusConflict, "not_deletable", "event not found or no longer a draft")
	}
	return nil
}

func (a *App) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return a.Delete(ctx, id)
}

func (a *App) ListFlashOlderThan(ctx context.Context, cutoff pgtype.Date) ([]dbgen.Event, error) {
	return a.ListFlashOlderThanRepo(ctx, cutoff)
}

type EventContext struct {
	Status    string
	EventType string
	StartDate pgtype.Date
	EndDate   pgtype.Date
	Dates     map[string]bool
}

func (c EventContext) IsDraft() bool { return c.Status == "draft" }

func (a *App) GetContext(ctx context.Context, orgID, eventID uuid.UUID) (EventContext, error) {
	event, err := a.GetEvent(ctx, orgID, eventID)
	if err != nil {
		return EventContext{}, utils.ErrNotFound("event")
	}
	eventDays, err := a.ListDays(ctx, eventID)
	if err != nil {
		return EventContext{}, utils.ErrInternal()
	}

	dates := make(map[string]bool, len(eventDays))
	for _, d := range eventDays {
		dates[timeutil.FormatDate(d.Date)] = true
	}
	return EventContext{
		Status:    event.Status,
		EventType: event.EventType,
		StartDate: event.StartDate,
		EndDate:   event.EndDate,
		Dates:     dates,
	}, nil
}

var dayCSVColumns = []string{"date", "entry_time", "exit_time"}

func (a *App) ImportDaysCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader) (DayImportSummary, error) {
	existing, err := a.GetEvent(ctx, orgID, eventID)
	if err != nil {
		return DayImportSummary{}, utils.ErrNotFound("event")
	}
	if existing.Status != "draft" {
		return DayImportSummary{}, utils.ErrConflict("only a draft event can be edited")
	}
	if existing.EventType != "standard" {
		return DayImportSummary{}, utils.ErrValidation(map[string]string{"event_type": "day import is only available for standard events"})
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return DayImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "could not read CSV header")
	}
	columns := make(map[string]int, len(header))
	for i, col := range header {
		columns[strings.ToLower(strings.TrimSpace(col))] = i
	}
	for _, required := range dayCSVColumns {
		if _, ok := columns[required]; !ok {
			return DayImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "missing required column: "+required)
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
		if date.Time.Before(existing.StartDate.Time) || date.Time.After(existing.EndDate.Time) {
			summary.Skipped++
			summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "date is outside the event's date range: " + dateStr})
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

	if len(days) == 0 {
		return DayImportSummary{}, utils.ErrValidation(map[string]string{"days": "no valid rows to import"})
	}
	minDate, maxDate := minMaxEventDates(days)

	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		if _, err := a.UpdateDates(ctx, tx, orgID, eventID, minDate, maxDate); err != nil {
			return utils.ErrConflict("only a draft event can be edited")
		}
		_, err := a.writeEventDaysTx(ctx, tx, eventID, days)
		return err
	})
	if txErr != nil {
		return DayImportSummary{}, txErr
	}

	summary.Imported = len(days)
	return summary, nil
}

func (a *App) ExportDays(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, error) {
	if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
		return nil, utils.ErrNotFound("event")
	}
	days, err := a.ListDays(ctx, eventID)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"date", "entry_time", "exit_time"})
	for _, d := range days {
		_ = w.Write([]string{timeutil.FormatDate(d.Date), timeutil.FormatClock(d.EntryTime), timeutil.FormatClock(d.ExitTime)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, utils.ErrInternal()
	}
	return buf.Bytes(), nil
}

func (a *App) UploadImage(ctx context.Context, orgID, eventID uuid.UUID, contentType string, size int64, r io.Reader) error {
	if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
		return utils.ErrNotFound("event")
	}
	return a.uploadEventImage(ctx, eventID, "event-images", contentType, size, r, func(key string) error {
		_, err := a.UpdateImage(ctx, eventID, key)
		return err
	})
}

func (a *App) GetImage(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, string, error) {
	if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
		return nil, "", utils.ErrNotFound("event")
	}
	meta, err := a.GetMetadata(ctx, eventID)
	if err != nil || !meta.ImageObjectKey.Valid {
		return nil, "", utils.ErrNotFound("image")
	}
	data, contentType, err := a.storage.Get(ctx, meta.ImageObjectKey.String)
	if err != nil {
		return nil, "", utils.ErrNotFound("image")
	}
	return data, contentType, nil
}

func (a *App) UploadOrganizerSignature(ctx context.Context, orgID, eventID uuid.UUID, contentType string, size int64, r io.Reader) error {
	if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
		return utils.ErrNotFound("event")
	}
	return a.uploadEventImage(ctx, eventID, "event-organizer-signatures", contentType, size, r, func(key string) error {
		_, err := a.UpdateOrganizerSignature(ctx, eventID, key)
		return err
	})
}

func (a *App) GetOrganizerSignature(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, string, error) {
	if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
		return nil, "", utils.ErrNotFound("event")
	}
	meta, err := a.GetMetadata(ctx, eventID)
	if err != nil || !meta.OrganizerSignatureObjectKey.Valid {
		return nil, "", utils.ErrNotFound("organizer signature")
	}
	data, contentType, err := a.storage.Get(ctx, meta.OrganizerSignatureObjectKey.String)
	if err != nil {
		return nil, "", utils.ErrNotFound("organizer signature")
	}
	return data, contentType, nil
}

func (a *App) uploadEventImage(ctx context.Context, eventID uuid.UUID, prefix, contentType string, size int64, r io.Reader, storeKey func(string) error) error {
	ext, ok := allowedEventImageContentTypes[contentType]
	if !ok {
		return utils.NewError(http.StatusUnsupportedMediaType, "unsupported_type", "image must be PNG or JPEG")
	}
	if size <= 0 || size > maxEventImageBytes {
		return utils.ErrValidation(map[string]string{"file": "file must be under 2MB"})
	}
	data, err := io.ReadAll(io.LimitReader(r, maxEventImageBytes+1))
	if err != nil {
		return utils.ErrInternal()
	}
	if len(data) > maxEventImageBytes {
		return utils.ErrValidation(map[string]string{"file": "file must be under 2MB"})
	}
	if _, format, err := image.DecodeConfig(bytes.NewReader(data)); err != nil || allowedEventImageContentTypes["image/"+format] == "" {
		return utils.ErrValidation(map[string]string{"file": "file is not a valid PNG or JPEG image"})
	}

	key := fmt.Sprintf("%s/%s.%s", prefix, eventID, ext)
	if err := a.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		return utils.ErrInternal()
	}
	if err := storeKey(key); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

func (a *App) writeEventDaysTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, days []parsedDay) ([]dbgen.EventDay, error) {
	if err := a.DeleteDaysForEvent(ctx, tx, eventID); err != nil {
		return nil, utils.ErrInternal()
	}
	out := make([]dbgen.EventDay, 0, len(days))
	for _, d := range days {
		day, err := a.CreateDay(ctx, tx, eventID, d.date, d.entryTime, d.exitTime)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		out = append(out, day)
	}
	return out, nil
}

type parsedDay struct {
	date      pgtype.Date
	entryTime pgtype.Time
	exitTime  pgtype.Time
}

func parseEventDays(inputs []EventDayInput) ([]parsedDay, error) {
	seen := make(map[string]bool, len(inputs))
	days := make([]parsedDay, len(inputs))

	for i, in := range inputs {
		if seen[in.Date] {
			return nil, utils.ErrValidation(map[string]string{"days": "duplicate date: " + in.Date})
		}
		seen[in.Date] = true

		date, err := timeutil.ParseDate(in.Date)
		if err != nil {
			return nil, utils.ErrValidation(map[string]string{"days": "invalid date: " + in.Date})
		}
		entry, exit, err := parseTimeWindow(in.EntryTime, in.ExitTime)
		if err != nil {
			return nil, utils.ErrValidation(map[string]string{"days": "invalid time on " + in.Date})
		}

		days[i] = parsedDay{date: date, entryTime: entry, exitTime: exit}
	}
	return days, nil
}

func minMaxEventDates(days []parsedDay) (pgtype.Date, pgtype.Date) {
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

func parseTimeWindow(entryStr, exitStr string) (pgtype.Time, pgtype.Time, error) {
	entry, err := timeutil.ParseClock(entryStr)
	if err != nil {
		return pgtype.Time{}, pgtype.Time{}, utils.ErrValidation(map[string]string{"entry_time": "invalid entry_time: " + entryStr})
	}
	exit, err := timeutil.ParseClock(exitStr)
	if err != nil {
		return pgtype.Time{}, pgtype.Time{}, utils.ErrValidation(map[string]string{"exit_time": "invalid exit_time: " + exitStr})
	}
	if exit.Microseconds <= entry.Microseconds {
		return pgtype.Time{}, pgtype.Time{}, utils.ErrValidation(map[string]string{"exit_time": "exit_time must be after entry_time"})
	}
	return entry, exit, nil
}

func toEventSummary(event dbgen.Event, meta dbgen.EventMetadatum) EventSummary {
	var publishedAt *string
	if event.PublishedAt.Valid {
		s := event.PublishedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		publishedAt = &s
	}
	return EventSummary{
		ID:                    event.ID,
		EventType:             event.EventType,
		Status:                event.Status,
		StartDate:             timeutil.FormatDate(event.StartDate),
		EndDate:               timeutil.FormatDate(event.EndDate),
		Name:                  meta.Name,
		Venue:                 pgTextPtr(meta.Venue),
		OrganizerName:         pgTextPtr(meta.OrganizerName),
		HasImage:              meta.ImageObjectKey.Valid,
		HasOrganizerSignature: meta.OrganizerSignatureObjectKey.Valid,
		PublishedAt:           publishedAt,
	}
}

func toEventResponse(event dbgen.Event, meta dbgen.EventMetadatum, eventDays []dbgen.EventDay) EventResponse {
	dayResponses := make([]EventDayResponse, len(eventDays))
	for i, d := range eventDays {
		dayResponses[i] = EventDayResponse{
			Date:      timeutil.FormatDate(d.Date),
			EntryTime: timeutil.FormatClock(d.EntryTime),
			ExitTime:  timeutil.FormatClock(d.ExitTime),
		}
	}
	return EventResponse{
		EventSummary: toEventSummary(event, meta),
		Days:         dayResponses,
	}
}

func textOrNull(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func pgTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
