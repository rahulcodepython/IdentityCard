package events

import (
    "bytes"
    "context"
    "encoding/csv"
    "errors"
    "fmt"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/generic"
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
    var startDate, endDate time.Time
    var err error

    switch req.EventType {
    case "flash":
        if len(req.Days) != 1 {
            return EventResponse{}, utils.ErrValidation(map[string]string{"days": "A flash event must have exactly one day."})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return EventResponse{}, err
        }
        startDate = days[0].date
        endDate = days[0].date

    case "standard":
        if len(req.Days) == 0 {
            return EventResponse{}, utils.ErrValidation(map[string]string{"days": "At least one day is required for a standard event."})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return EventResponse{}, err
        }
        startDate, endDate = minMaxEventDates(days)

    case "grouped":
        if req.RangeStart == "" || req.RangeEnd == "" {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "range_start and range_end are required for a grouped event."})
        }
        pgStart, err := timeutil.ParseDate(req.RangeStart)
        if err != nil {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "Invalid start date."})
        }
        pgEnd, err := timeutil.ParseDate(req.RangeEnd)
        if err != nil {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "Invalid end date."})
        }
        startDate = pgStart.Time
        endDate = pgEnd.Time
        if endDate.Before(startDate) {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "range_end must be on or after range_start."})
        }

    default:
        return EventResponse{}, utils.ErrValidation(map[string]string{"event_type": "Unknown event type."})
    }

    var eventID uuid.UUID
    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        credit, err := a.plans.FindCreditTx(ctx, tx, orgID, req.EventType)
        if err != nil {
            return err
        }

        createdEvent, err := a.CreateEvent(ctx, tx, orgID, req.EventType, startDate, endDate)
        if err != nil {
            return err
        }
        eventID = createdEvent.ID

        if len(days) > 0 {
            if _, err := a.writeEventDaysTx(ctx, tx, eventID, days); err != nil {
                return err
            }
        }

        if _, err := a.CreateMetadata(ctx, tx, eventID, req.Name, strPtrOrNil(req.Venue), strPtrOrNil(req.OrganizerName)); err != nil {
            return err
        }

        if err := a.plans.LinkCreditTx(ctx, tx, credit.ID, eventID); err != nil {
            return err
        }

        return a.plans.MaybeReplenishUnlimitedTx(ctx, tx, orgID, credit.Type)
    })
    if txErr != nil {
        if errors.Is(txErr, generic.ErrPlansNoCredits) {
            return EventResponse{}, utils.NewError(http.StatusForbidden, "No available credit for this event type — purchase a plan or wait for renewal.", txErr)
        }
        var apiErr *utils.APIError
        if errors.As(txErr, &apiErr) {
            return EventResponse{}, apiErr
        }
        return EventResponse{}, utils.ErrInternal("Failed to create event.", txErr)
    }

    resp, err := a.Get(ctx, orgID, eventID)
    if err != nil {
        return EventResponse{}, err
    }
    return resp, nil
}

func (a *App) Get(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
    event, err := a.GetEvent(ctx, orgID, id)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrEventsNotFound) {
            return EventResponse{}, utils.ErrNotFound("Event not found.", err)
        }
        return EventResponse{}, utils.ErrInternal("Failed to fetch event.", err)
    }
    return *event, nil
}

func (a *App) List(ctx context.Context, orgID uuid.UUID) ([]EventSummary, error) {
    rows, err := a.ListEvents(ctx, orgID)
    if err != nil {
        return nil, utils.ErrInternal("Failed to list events.", err)
    }
    return rows, nil
}

func (a *App) Update(ctx context.Context, orgID, id uuid.UUID, req UpdateEventRequest) (EventResponse, error) {
    existing, err := a.GetEvent(ctx, orgID, id)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrEventsNotFound) {
            return EventResponse{}, utils.ErrNotFound("Event not found.", err)
        }
        return EventResponse{}, utils.ErrInternal("Failed to fetch event.", err)
    }
    if existing.Status != "draft" {
        return EventResponse{}, utils.ErrConflict("Only a draft event can be edited.", generic.ErrEventsNotDraft)
    }

    restricted, err := a.plans.GetCreditRestriction(ctx, id)
    if err != nil {
        return EventResponse{}, utils.ErrInternal("Failed to check credit restriction.", err)
    }
    if restricted {
        return EventResponse{}, utils.NewError(http.StatusForbidden, "This event's billing is past due — renew your plan to edit.", nil)
    }

    var days []parsedDay
    var startDate, endDate time.Time

    switch existing.EventType {
    case "flash":
        if len(req.Days) != 1 {
            return EventResponse{}, utils.ErrValidation(map[string]string{"days": "A flash event must have exactly one day."})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return EventResponse{}, err
        }
        startDate = days[0].date
        endDate = days[0].date

    case "standard":
        if len(req.Days) == 0 {
            return EventResponse{}, utils.ErrValidation(map[string]string{"days": "At least one day is required for a standard event."})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return EventResponse{}, err
        }
        startDate, endDate = minMaxEventDates(days)

    case "grouped":
        if req.RangeStart == "" || req.RangeEnd == "" {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "range_start and range_end are required for a grouped event."})
        }
        pgStart, err := timeutil.ParseDate(req.RangeStart)
        if err != nil {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "Invalid start date."})
        }
        pgEnd, err := timeutil.ParseDate(req.RangeEnd)
        if err != nil {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "Invalid end date."})
        }
        startDate = pgStart.Time
        endDate = pgEnd.Time
        if endDate.Before(startDate) {
            return EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "range_end must be on or after range_start."})
        }
    }

    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        if _, err := a.UpdateDates(ctx, tx, orgID, id, startDate, endDate); err != nil {
            return err
        }

        if _, err := a.UpdateMetadata(ctx, tx, id, req.Name, strPtrOrNil(req.Venue), strPtrOrNil(req.OrganizerName)); err != nil {
            return err
        }

        if len(days) > 0 {
            if _, err := a.writeEventDaysTx(ctx, tx, id, days); err != nil {
                return err
            }
        }
        return nil
    })
    if txErr != nil {
        return EventResponse{}, utils.ErrInternal("Failed to update event.", txErr)
    }

    return a.Get(ctx, orgID, id)
}

func (a *App) Publish(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        published, err := a.PublishEvent(ctx, tx, orgID, id)
        if err != nil {
            return utils.NewError(http.StatusConflict, "Event not found or already published.", err)
        }
        pubTime := time.Now().UTC()
        if published.PublishedAt != nil {
            pubTime = *published.PublishedAt
        }
        return a.MarkUnjoinedPeopleJoinedAt(ctx, tx, id, pubTime)
    })
    if txErr != nil {
        return EventResponse{}, utils.ErrInternal("Failed to publish event.", txErr)
    }
    return a.Get(ctx, orgID, id)
}

func (a *App) DeleteDraftEvent(ctx context.Context, orgID, id uuid.UUID) error {
    deleted, err := a.DeleteDraft(ctx, orgID, id)
    if err != nil {
        return utils.ErrInternal("Failed to delete event.", err)
    }
    if !deleted {
        return utils.NewError(http.StatusConflict, "Event not found or no longer a draft.", generic.ErrEventsNotDraft)
    }
    return nil
}

func (a *App) DeleteByID(ctx context.Context, id uuid.UUID) error {
    return a.Delete(ctx, id)
}

func (a *App) ListFlashOlderThan(ctx context.Context, cutoff time.Time) ([]EventDB, error) {
    return a.ListFlashOlderThanRepo(ctx, cutoff)
}

type EventContext struct {
    Status    string
    EventType string
    StartDate string
    EndDate   string
    Dates     map[string]bool
}

func (c EventContext) IsDraft() bool { return c.Status == "draft" }

func (a *App) GetContext(ctx context.Context, orgID, eventID uuid.UUID) (EventContext, error) {
    event, err := a.GetEvent(ctx, orgID, eventID)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrEventsNotFound) {
            return EventContext{}, utils.ErrNotFound("Event not found.", err)
        }
        return EventContext{}, utils.ErrInternal("Failed to fetch event context.", err)
    }
    eventDays, err := a.ListDays(ctx, eventID)
    if err != nil {
        return EventContext{}, utils.ErrInternal("Failed to list event days.", err)
    }

    dates := make(map[string]bool, len(eventDays))
    for _, d := range eventDays {
        dates[d.Date] = true
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
        return DayImportSummary{}, utils.ErrNotFound("Event not found.", err)
    }
    if existing.Status != "draft" {
        return DayImportSummary{}, utils.ErrConflict("Only a draft event can be edited.", generic.ErrEventsNotDraft)
    }
    if existing.EventType != "standard" {
        return DayImportSummary{}, utils.ErrValidation(map[string]string{"event_type": "Day import is only available for standard events."})
    }

    reader := csv.NewReader(file)
    reader.TrimLeadingSpace = true
    header, err := reader.Read()
    if err != nil {
        return DayImportSummary{}, utils.NewError(http.StatusBadRequest, "Could not read CSV header.", err)
    }
    columns := make(map[string]int, len(header))
    for i, col := range header {
        columns[strings.ToLower(strings.TrimSpace(col))] = i
    }
    for _, required := range dayCSVColumns {
        if _, ok := columns[required]; !ok {
            return DayImportSummary{}, utils.NewError(http.StatusBadRequest, "Missing required column: "+required, nil)
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
            summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "Malformed row."})
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
            summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "Missing date, entry_time, or exit_time."})
            continue
        }
        if seen[dateStr] {
            summary.Skipped++
            summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "Duplicate date: " + dateStr})
            continue
        }
        pgDate, err := timeutil.ParseDate(dateStr)
        if err != nil {
            summary.Skipped++
            summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "Invalid date: " + dateStr})
            continue
        }
        date := pgDate.Time
        entry, exit, err := parseTimeWindow(entryStr, exitStr)
        if err != nil {
            summary.Skipped++
            summary.Errors = append(summary.Errors, DayImportRowError{Row: rowNum, Message: "Invalid entry/exit time."})
            continue
        }
        seen[dateStr] = true
        days = append(days, parsedDay{date: date, entryTime: entry, exitTime: exit})
    }

    if len(days) == 0 {
        return DayImportSummary{}, utils.ErrValidation(map[string]string{"days": "No valid rows to import."})
    }
    minDate, maxDate := minMaxEventDates(days)

    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        if _, err := a.UpdateDates(ctx, tx, orgID, eventID, minDate, maxDate); err != nil {
            return err
        }
        _, err := a.writeEventDaysTx(ctx, tx, eventID, days)
        return err
    })
    if txErr != nil {
        return DayImportSummary{}, utils.ErrInternal("Failed to save imported days.", txErr)
    }

    summary.Imported = len(days)
    return summary, nil
}

func (a *App) ExportDays(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, error) {
    if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
        return nil, utils.ErrNotFound("Event not found.", err)
    }
    days, err := a.ListDays(ctx, eventID)
    if err != nil {
        return nil, utils.ErrInternal("Failed to list event days.", err)
    }

    var buf bytes.Buffer
    w := csv.NewWriter(&buf)
    _ = w.Write([]string{"date", "entry_time", "exit_time"})
    for _, d := range days {
        _ = w.Write([]string{d.Date, d.EntryTime, d.ExitTime})
    }
    w.Flush()
    if err := w.Error(); err != nil {
        return nil, utils.ErrInternal("Failed to export days CSV.", err)
    }
    return buf.Bytes(), nil
}

func (a *App) UploadImage(ctx context.Context, orgID, eventID uuid.UUID, contentType string, size int64, r io.Reader) error {
    if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
        return utils.ErrNotFound("Event not found.", err)
    }
    return a.uploadEventImage(ctx, eventID, "event-images", contentType, size, r, func(key string) error {
        _, err := a.UpdateImage(ctx, eventID, key)
        return err
    })
}

func (a *App) GetImage(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, string, error) {
    if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
        return nil, "", utils.ErrNotFound("Event not found.", err)
    }
    meta, err := a.GetMetadata(ctx, eventID)
    if err != nil || meta.ImageObjectKey == nil || *meta.ImageObjectKey == "" {
        return nil, "", utils.ErrNotFound("Event image not found.", err)
    }
    data, contentType, err := a.storage.Get(ctx, *meta.ImageObjectKey)
    if err != nil {
        return nil, "", utils.ErrNotFound("Image file not found in storage.", err)
    }
    return data, contentType, nil
}

func (a *App) UploadOrganizerSignature(ctx context.Context, orgID, eventID uuid.UUID, contentType string, size int64, r io.Reader) error {
    if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
        return utils.ErrNotFound("Event not found.", err)
    }
    return a.uploadEventImage(ctx, eventID, "event-organizer-signatures", contentType, size, r, func(key string) error {
        _, err := a.UpdateOrganizerSignature(ctx, eventID, key)
        return err
    })
}

func (a *App) GetOrganizerSignature(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, string, error) {
    if _, err := a.GetEvent(ctx, orgID, eventID); err != nil {
        return nil, "", utils.ErrNotFound("Event not found.", err)
    }
    meta, err := a.GetMetadata(ctx, eventID)
    if err != nil || meta.OrganizerSignatureObjectKey == nil || *meta.OrganizerSignatureObjectKey == "" {
        return nil, "", utils.ErrNotFound("Organizer signature not found.", err)
    }
    data, contentType, err := a.storage.Get(ctx, *meta.OrganizerSignatureObjectKey)
    if err != nil {
        return nil, "", utils.ErrNotFound("Signature file not found in storage.", err)
    }
    return data, contentType, nil
}

func (a *App) uploadEventImage(ctx context.Context, eventID uuid.UUID, prefix, contentType string, size int64, r io.Reader, storeKey func(string) error) error {
    ext, ok := allowedEventImageContentTypes[contentType]
    if !ok {
        return utils.NewError(http.StatusUnsupportedMediaType, "Image must be PNG or JPEG.", nil)
    }
    if size <= 0 || size > maxEventImageBytes {
        return utils.ErrValidation(map[string]string{"file": "File must be under 2MB."})
    }
    data, err := io.ReadAll(io.LimitReader(r, maxEventImageBytes+1))
    if err != nil {
        return utils.ErrInternal("Failed to read image data.", err)
    }
    if len(data) > maxEventImageBytes {
        return utils.ErrValidation(map[string]string{"file": "File must be under 2MB."})
    }
    if _, format, err := image.DecodeConfig(bytes.NewReader(data)); err != nil || allowedEventImageContentTypes["image/"+format] == "" {
        return utils.ErrValidation(map[string]string{"file": "File is not a valid PNG or JPEG image."})
    }

    key := fmt.Sprintf("%s/%s.%s", prefix, eventID, ext)
    if err := a.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
        return utils.ErrInternal("Failed to store image file.", err)
    }
    if err := storeKey(key); err != nil {
        return utils.ErrInternal("Failed to update image reference.", err)
    }
    return nil
}

func (a *App) writeEventDaysTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, days []parsedDay) ([]EventDayDB, error) {
    if err := a.DeleteDaysForEvent(ctx, tx, eventID); err != nil {
        return nil, err
    }
    out := make([]EventDayDB, 0, len(days))
    for _, d := range days {
        day, err := a.CreateDay(ctx, tx, eventID, d.date, d.entryTime, d.exitTime)
        if err != nil {
            return nil, err
        }
        out = append(out, *day)
    }
    return out, nil
}

type parsedDay struct {
    date      time.Time
    entryTime string
    exitTime  string
}

func parseEventDays(inputs []EventDayInput) ([]parsedDay, error) {
    seen := make(map[string]bool, len(inputs))
    days := make([]parsedDay, len(inputs))

    for i, in := range inputs {
        if seen[in.Date] {
            return nil, utils.ErrValidation(map[string]string{"days": "Duplicate date: " + in.Date})
        }
        seen[in.Date] = true

        pgDate, err := timeutil.ParseDate(in.Date)
        if err != nil {
            return nil, utils.ErrValidation(map[string]string{"days": "Invalid date: " + in.Date})
        }
        entry, exit, err := parseTimeWindow(in.EntryTime, in.ExitTime)
        if err != nil {
            return nil, utils.ErrValidation(map[string]string{"days": "Invalid time window on " + in.Date})
        }

        days[i] = parsedDay{date: pgDate.Time, entryTime: entry, exitTime: exit}
    }
    return days, nil
}

func minMaxEventDates(days []parsedDay) (time.Time, time.Time) {
    min, max := days[0].date, days[0].date
    for _, d := range days[1:] {
        if d.date.Before(min) {
            min = d.date
        }
        if d.date.After(max) {
            max = d.date
        }
    }
    return min, max
}

func parseTimeWindow(entryStr, exitStr string) (string, string, error) {
    entry, err := timeutil.ParseClock(entryStr)
    if err != nil {
        return "", "", utils.ErrValidation(map[string]string{"entry_time": "Invalid entry_time: " + entryStr})
    }
    exit, err := timeutil.ParseClock(exitStr)
    if err != nil {
        return "", "", utils.ErrValidation(map[string]string{"exit_time": "Invalid exit_time: " + exitStr})
    }
    if exit.Microseconds <= entry.Microseconds {
        return "", "", utils.ErrValidation(map[string]string{"exit_time": "exit_time must be after entry_time"})
    }
    return entryStr, exitStr, nil
}

func strPtrOrNil(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
