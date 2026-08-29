package services

import (
    "bytes"
    "context"
    "encoding/csv"
    "fmt"
    "io"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "net/http"
    "strings"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/jackc/pgx/v5/pgxpool"

    dbgen "identitycard-server/internal/db/sqlc/generated"
    "identitycard-server/internal/entities"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/pkg/storage"
    "identitycard-server/internal/repositories"
    "identitycard-server/internal/utils"
    "identitycard-server/internal/utils/timeutil"
)

// maxEventImageBytes caps image uploads at 2 MB — matches the org logo limit.
const maxEventImageBytes = 2 << 20

var allowedEventImageContentTypes = map[string]string{
    "image/png":  "png",
    "image/jpeg": "jpg",
}

type EventsService struct {
    repo    *repositories.EventsRepository
    pool    *pgxpool.Pool
    plans   *PlansService
    storage *storage.Storage
}

func NewEventsService(
    repo *repositories.EventsRepository,
    pool *pgxpool.Pool,
    plansService *PlansService,
    objectStore *storage.Storage,
) *EventsService {
    return &EventsService{repo: repo, pool: pool, plans: plansService, storage: objectStore}
}

// Create validates the request per event_type, consumes a credit inside a
// single transaction, writes the event + event_days (flash/standard only) +
// event_metadata row atomically, and returns the full EventResponse.
func (s *EventsService) Create(ctx context.Context, orgID uuid.UUID, req entities.CreateEventRequest) (entities.EventResponse, error) {
    var days []parsedDay
    var startDate, endDate pgtype.Date
    var err error

    switch req.EventType {
    case "flash":
        if len(req.Days) != 1 {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return entities.EventResponse{}, err
        }
        startDate = days[0].date
        endDate = days[0].date

    case "standard":
        if len(req.Days) == 0 {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"days": "at least one day is required for a standard event"})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return entities.EventResponse{}, err
        }
        startDate, endDate = minMaxEventDates(days)

    case "grouped":
        if req.RangeStart == "" || req.RangeEnd == "" {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "range_start and range_end are required for a grouped event"})
        }
        startDate, err = timeutil.ParseDate(req.RangeStart)
        if err != nil {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "invalid date"})
        }
        endDate, err = timeutil.ParseDate(req.RangeEnd)
        if err != nil {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "invalid date"})
        }
        if endDate.Time.Before(startDate.Time) {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "range_end must be on or after range_start"})
        }
        // days stays nil — grouped events have no event_days rows

    default:
        return entities.EventResponse{}, utils.ErrValidation(map[string]string{"event_type": "unknown event type"})
    }

    var event dbgen.Event
    var meta dbgen.EventMetadatum
    var createdDays []dbgen.EventDay

    txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
        // 1. Find + lock a credit before writing anything.
        credit, err := s.plans.FindCreditTx(ctx, tx, orgID, req.EventType)
        if err != nil {
            return err
        }

        // 2. Insert the event row.
        repo := s.repo.WithTx(tx)
        event, err = repo.Create(ctx, orgID, req.EventType, startDate, endDate)
        if err != nil {
            return utils.ErrInternal()
        }

        // 3. Write event_days (flash/standard only).
        if len(days) > 0 {
            createdDays, err = writeEventDaysTx(ctx, repo, event.ID, days)
            if err != nil {
                return err
            }
        }

        // 4. Write event_metadata (name/venue/organizer_name).
        meta, err = repo.CreateMetadata(ctx, event.ID, req.Name, textOrNull(req.Venue), textOrNull(req.OrganizerName))
        if err != nil {
            return utils.ErrInternal()
        }

        // 5. Link the credit to this event.
        if err := s.plans.LinkCreditTx(ctx, tx, credit.ID, event.ID); err != nil {
            return err
        }

        // 6. Replenish unlimited credits if needed.
        return s.plans.MaybeReplenishUnlimitedTx(ctx, tx, orgID, credit.Type)
    })
    if txErr != nil {
        return entities.EventResponse{}, txErr
    }

    return toEventResponse(event, meta, createdDays), nil
}

func (s *EventsService) Get(ctx context.Context, orgID, id uuid.UUID) (entities.EventResponse, error) {
    event, err := s.repo.Get(ctx, orgID, id)
    if err != nil {
        return entities.EventResponse{}, utils.ErrNotFound("event")
    }
    meta, err := s.repo.GetMetadata(ctx, id)
    if err != nil {
        return entities.EventResponse{}, utils.ErrInternal()
    }
    eventDays, err := s.repo.ListDays(ctx, id)
    if err != nil {
        return entities.EventResponse{}, utils.ErrInternal()
    }
    return toEventResponse(event, meta, eventDays), nil
}

func (s *EventsService) List(ctx context.Context, orgID uuid.UUID) ([]entities.EventSummary, error) {
    rows, err := s.repo.List(ctx, orgID)
    if err != nil {
        return nil, utils.ErrInternal()
    }
    summaries := make([]entities.EventSummary, len(rows))
    for i, row := range rows {
        meta, err := s.repo.GetMetadata(ctx, row.ID)
        if err != nil {
            // Metadata is always created together with the event; treat a
            // missing row as an internal inconsistency rather than 404.
            return nil, utils.ErrInternal()
        }
        summaries[i] = toEventSummary(row, meta)
    }
    return summaries, nil
}

// Update allows editing a draft event's metadata (name/venue/organizer_name)
// and — for flash/standard — its day list. Blocked if the funding credit is
// restricted (billing lapsed). event_type is immutable.
func (s *EventsService) Update(ctx context.Context, orgID, id uuid.UUID, req entities.UpdateEventRequest) (entities.EventResponse, error) {
    existing, err := s.repo.Get(ctx, orgID, id)
    if err != nil {
        return entities.EventResponse{}, utils.ErrNotFound("event")
    }
    if existing.Status != "draft" {
        return entities.EventResponse{}, utils.ErrConflict("only a draft event can be edited")
    }

    // Gate on credit restriction (billing lapsed).
    restricted, err := s.plans.GetCreditRestriction(ctx, id)
    if err != nil {
        return entities.EventResponse{}, err
    }
    if restricted {
        return entities.EventResponse{}, utils.NewError(http.StatusForbidden, "credit_restricted", "this event's billing is past due — renew your plan to edit")
    }

    var days []parsedDay
    var startDate, endDate pgtype.Date

    switch existing.EventType {
    case "flash":
        if len(req.Days) != 1 {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return entities.EventResponse{}, err
        }
        startDate = days[0].date
        endDate = days[0].date

    case "standard":
        if len(req.Days) == 0 {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"days": "at least one day is required for a standard event"})
        }
        days, err = parseEventDays(req.Days)
        if err != nil {
            return entities.EventResponse{}, err
        }
        startDate, endDate = minMaxEventDates(days)

    case "grouped":
        if req.RangeStart == "" || req.RangeEnd == "" {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "range_start and range_end are required for a grouped event"})
        }
        startDate, err = timeutil.ParseDate(req.RangeStart)
        if err != nil {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_start": "invalid date"})
        }
        endDate, err = timeutil.ParseDate(req.RangeEnd)
        if err != nil {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "invalid date"})
        }
        if endDate.Time.Before(startDate.Time) {
            return entities.EventResponse{}, utils.ErrValidation(map[string]string{"range_end": "range_end must be on or after range_start"})
        }
    }

    var meta dbgen.EventMetadatum
    var updatedDays []dbgen.EventDay

    txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
        repo := s.repo.WithTx(tx)

        // Update the structural date range.
        if _, err := repo.UpdateDates(ctx, orgID, id, startDate, endDate); err != nil {
            return utils.ErrConflict("only a draft event can be edited")
        }

        // Update metadata.
        meta, err = repo.UpdateMetadata(ctx, id, req.Name, textOrNull(req.Venue), textOrNull(req.OrganizerName))
        if err != nil {
            return utils.ErrInternal()
        }

        // Replace event_days for flash/standard.
        if len(days) > 0 {
            updatedDays, err = writeEventDaysTx(ctx, repo, id, days)
            if err != nil {
                return err
            }
        }
        return nil
    })
    if txErr != nil {
        return entities.EventResponse{}, txErr
    }

    return toEventResponse(existing, meta, updatedDays), nil
}

func (s *EventsService) Publish(ctx context.Context, orgID, id uuid.UUID) (entities.EventResponse, error) {
    txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
        repo := s.repo.WithTx(tx)
        published, err := repo.Publish(ctx, orgID, id)
        if err != nil {
            return utils.NewError(http.StatusConflict, "not_publishable", "event not found or already published")
        }
        return repo.MarkUnjoinedPeopleJoinedAt(ctx, id, published.PublishedAt)
    })
    if txErr != nil {
        return entities.EventResponse{}, txErr
    }
    return s.Get(ctx, orgID, id)
}

// Delete removes a draft event scoped to orgID — gated on draft status, user-facing.
func (s *EventsService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
    deleted, err := s.repo.DeleteDraft(ctx, orgID, id)
    if err != nil {
        return utils.ErrInternal()
    }
    if !deleted {
        return utils.NewError(http.StatusConflict, "not_deletable", "event not found or no longer a draft")
    }
    return nil
}

// DeleteByID unconditionally hard-deletes an event by ID — system-only, used by
// the cron cleanup job. Not gated on draft status or org scope.
func (s *EventsService) DeleteByID(ctx context.Context, id uuid.UUID) error {
    return s.repo.Delete(ctx, id)
}

// ListFlashOlderThan returns flash events whose single day is older than the
// cutoff — used by the daily cleanup cron.
func (s *EventsService) ListFlashOlderThan(ctx context.Context, cutoff pgtype.Date) ([]dbgen.Event, error) {
    return s.repo.ListFlashOlderThan(ctx, cutoff)
}

// EventContext is the minimal parent-event read shared by every domain nested
// under an event (sub-events, people, forms): is it still editable, and which
// dates are valid children (empty for grouped events — sub-events supply their
// own date, validated against the parent's [start_date, end_date] range instead).
type EventContext struct {
    Status    string
    EventType string
    StartDate pgtype.Date
    EndDate   pgtype.Date
    Dates     map[string]bool // populated for flash/standard; empty for grouped
}

func (c EventContext) IsDraft() bool { return c.Status == "draft" }

func (s *EventsService) GetContext(ctx context.Context, orgID, eventID uuid.UUID) (EventContext, error) {
    event, err := s.repo.Get(ctx, orgID, eventID)
    if err != nil {
        return EventContext{}, utils.ErrNotFound("event")
    }
    eventDays, err := s.repo.ListDays(ctx, eventID)
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

// ImportDaysCSV wholesale-replaces a draft standard event's day list from a
// CSV upload — same replace semantics as Update. Each date is validated to
// fall within [start_date, end_date] and not be duplicated; bad rows are
// skipped and reported.
func (s *EventsService) ImportDaysCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader) (entities.DayImportSummary, error) {
    existing, err := s.repo.Get(ctx, orgID, eventID)
    if err != nil {
        return entities.DayImportSummary{}, utils.ErrNotFound("event")
    }
    if existing.Status != "draft" {
        return entities.DayImportSummary{}, utils.ErrConflict("only a draft event can be edited")
    }
    if existing.EventType != "standard" {
        return entities.DayImportSummary{}, utils.ErrValidation(map[string]string{"event_type": "day import is only available for standard events"})
    }

    reader := csv.NewReader(file)
    reader.TrimLeadingSpace = true
    header, err := reader.Read()
    if err != nil {
        return entities.DayImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "could not read CSV header")
    }
    columns := make(map[string]int, len(header))
    for i, col := range header {
        columns[strings.ToLower(strings.TrimSpace(col))] = i
    }
    for _, required := range dayCSVColumns {
        if _, ok := columns[required]; !ok {
            return entities.DayImportSummary{}, utils.NewError(http.StatusBadRequest, "bad_csv", "missing required column: "+required)
        }
    }

    summary := entities.DayImportSummary{}
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
            summary.Errors = append(summary.Errors, entities.DayImportRowError{Row: rowNum, Message: "malformed row"})
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
            summary.Errors = append(summary.Errors, entities.DayImportRowError{Row: rowNum, Message: "missing date, entry_time, or exit_time"})
            continue
        }
        if seen[dateStr] {
            summary.Skipped++
            summary.Errors = append(summary.Errors, entities.DayImportRowError{Row: rowNum, Message: "duplicate date: " + dateStr})
            continue
        }
        date, err := timeutil.ParseDate(dateStr)
        if err != nil {
            summary.Skipped++
            summary.Errors = append(summary.Errors, entities.DayImportRowError{Row: rowNum, Message: "invalid date"})
            continue
        }
        // Validate the date falls within the event's own [start_date, end_date].
        if date.Time.Before(existing.StartDate.Time) || date.Time.After(existing.EndDate.Time) {
            summary.Skipped++
            summary.Errors = append(summary.Errors, entities.DayImportRowError{Row: rowNum, Message: "date is outside the event's date range: " + dateStr})
            continue
        }
        entry, exit, err := parseTimeWindow(entryStr, exitStr)
        if err != nil {
            summary.Skipped++
            summary.Errors = append(summary.Errors, entities.DayImportRowError{Row: rowNum, Message: "invalid entry_time/exit_time"})
            continue
        }
        seen[dateStr] = true
        days = append(days, parsedDay{date: date, entryTime: entry, exitTime: exit})
    }

    if len(days) == 0 {
        return entities.DayImportSummary{}, utils.ErrValidation(map[string]string{"days": "no valid rows to import"})
    }
    minDate, maxDate := minMaxEventDates(days)

    txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
        repo := s.repo.WithTx(tx)
        if _, err := repo.UpdateDates(ctx, orgID, eventID, minDate, maxDate); err != nil {
            return utils.ErrConflict("only a draft event can be edited")
        }
        _, err := writeEventDaysTx(ctx, repo, eventID, days)
        return err
    })
    if txErr != nil {
        return entities.DayImportSummary{}, txErr
    }

    summary.Imported = len(days)
    return summary, nil
}

// ExportDays returns a CSV of the event's event_days — always reads from the
// materialized rows regardless of event_type.
func (s *EventsService) ExportDays(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, error) {
    if _, err := s.repo.Get(ctx, orgID, eventID); err != nil {
        return nil, utils.ErrNotFound("event")
    }
    days, err := s.repo.ListDays(ctx, eventID)
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

// UploadImage stores an event image in MinIO and records its object key —
// mirrors OrganizationsService.UploadLogo.
func (s *EventsService) UploadImage(ctx context.Context, orgID, eventID uuid.UUID, contentType string, size int64, r io.Reader) error {
    // Verify the event belongs to this org.
    if _, err := s.repo.Get(ctx, orgID, eventID); err != nil {
        return utils.ErrNotFound("event")
    }
    return s.uploadEventImage(ctx, eventID, "event-images", contentType, size, r, func(key string) error {
        _, err := s.repo.UpdateImage(ctx, eventID, key)
        return err
    })
}

// GetImage retrieves the event's image bytes from MinIO.
func (s *EventsService) GetImage(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, string, error) {
    if _, err := s.repo.Get(ctx, orgID, eventID); err != nil {
        return nil, "", utils.ErrNotFound("event")
    }
    meta, err := s.repo.GetMetadata(ctx, eventID)
    if err != nil || !meta.ImageObjectKey.Valid {
        return nil, "", utils.ErrNotFound("image")
    }
    data, contentType, err := s.storage.Get(ctx, meta.ImageObjectKey.String)
    if err != nil {
        return nil, "", utils.ErrNotFound("image")
    }
    return data, contentType, nil
}

// UploadOrganizerSignature stores the organizer's signature image in MinIO.
func (s *EventsService) UploadOrganizerSignature(ctx context.Context, orgID, eventID uuid.UUID, contentType string, size int64, r io.Reader) error {
    if _, err := s.repo.Get(ctx, orgID, eventID); err != nil {
        return utils.ErrNotFound("event")
    }
    return s.uploadEventImage(ctx, eventID, "event-organizer-signatures", contentType, size, r, func(key string) error {
        _, err := s.repo.UpdateOrganizerSignature(ctx, eventID, key)
        return err
    })
}

// GetOrganizerSignature retrieves the organizer signature image bytes.
func (s *EventsService) GetOrganizerSignature(ctx context.Context, orgID, eventID uuid.UUID) ([]byte, string, error) {
    if _, err := s.repo.Get(ctx, orgID, eventID); err != nil {
        return nil, "", utils.ErrNotFound("event")
    }
    meta, err := s.repo.GetMetadata(ctx, eventID)
    if err != nil || !meta.OrganizerSignatureObjectKey.Valid {
        return nil, "", utils.ErrNotFound("organizer signature")
    }
    data, contentType, err := s.storage.Get(ctx, meta.OrganizerSignatureObjectKey.String)
    if err != nil {
        return nil, "", utils.ErrNotFound("organizer signature")
    }
    return data, contentType, nil
}

// uploadEventImage is the shared MinIO upload helper for event images and
// organizer signatures — validates content type, size, and image decodability,
// then calls storeKey to persist the resulting object key.
func (s *EventsService) uploadEventImage(ctx context.Context, eventID uuid.UUID, prefix, contentType string, size int64, r io.Reader, storeKey func(string) error) error {
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
    if err := s.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
        return utils.ErrInternal()
    }
    if err := storeKey(key); err != nil {
        return utils.ErrInternal()
    }
    return nil
}

// -- shared per-tx writers --

func writeEventDaysTx(ctx context.Context, repo *repositories.EventsRepository, eventID uuid.UUID, days []parsedDay) ([]dbgen.EventDay, error) {
    if err := repo.DeleteDaysForEvent(ctx, eventID); err != nil {
        return nil, utils.ErrInternal()
    }
    out := make([]dbgen.EventDay, 0, len(days))
    for _, d := range days {
        day, err := repo.CreateDay(ctx, eventID, d.date, d.entryTime, d.exitTime)
        if err != nil {
            return nil, utils.ErrInternal()
        }
        out = append(out, day)
    }
    return out, nil
}

// -- explicit-day parsing (flash/standard) --

type parsedDay struct {
    date      pgtype.Date
    entryTime pgtype.Time
    exitTime  pgtype.Time
}

func parseEventDays(inputs []entities.EventDayInput) ([]parsedDay, error) {
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

// -- response/wire conversion --

func toEventSummary(event dbgen.Event, meta dbgen.EventMetadatum) entities.EventSummary {
    var publishedAt *string
    if event.PublishedAt.Valid {
        s := event.PublishedAt.Time.Format("2006-01-02T15:04:05Z07:00")
        publishedAt = &s
    }
    return entities.EventSummary{
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

func toEventResponse(event dbgen.Event, meta dbgen.EventMetadatum, eventDays []dbgen.EventDay) entities.EventResponse {
    dayResponses := make([]entities.EventDayResponse, len(eventDays))
    for i, d := range eventDays {
        dayResponses[i] = entities.EventDayResponse{
            Date:      timeutil.FormatDate(d.Date),
            EntryTime: timeutil.FormatClock(d.EntryTime),
            ExitTime:  timeutil.FormatClock(d.ExitTime),
        }
    }
    return entities.EventResponse{
        EventSummary: toEventSummary(event, meta),
        Days:         dayResponses,
    }
}

// textOrNull converts an empty string to a null pgtype.Text.
func textOrNull(s string) pgtype.Text {
    if s == "" {
        return pgtype.Text{}
    }
    return pgtype.Text{String: s, Valid: true}
}

// pgTextPtr returns nil if the pgtype.Text is not valid, or a *string otherwise.
func pgTextPtr(t pgtype.Text) *string {
    if !t.Valid {
        return nil
    }
    return &t.String
}

// todayDate returns today's date at UTC midnight — used by plans (same
// package) and cron jobs. Only one definition; plans.service.go owns it.
