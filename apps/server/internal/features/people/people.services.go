package people

import (
    "bytes"
    "context"
    "encoding/csv"
    "errors"
    "io"
    "net/http"
    "net/mail"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/utils"
)

func (a *App) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreatePersonRequest) (PersonResponse, error) {
    parent, err := a.events.GetContext(ctx, orgID, eventID)
    if err != nil {
        return PersonResponse{}, err
    }

    resp, _, err := a.upsert(ctx, orgID, eventID, peopleUpsertRequest{
        Email: req.Email, Mobile: req.Mobile, Name: req.Name,
        ImageURL: req.ImageURL, Age: req.Age, Gender: req.Gender,
        SubEventIDs: req.SubEventIDs,
    }, !parent.IsDraft())
    return resp, err
}

func (a *App) Submit(ctx context.Context, orgID, eventID uuid.UUID, in PeopleSubmitInput) (PersonResponse, bool, error) {
    parent, err := a.events.GetContext(ctx, orgID, eventID)
    if err != nil {
        return PersonResponse{}, false, err
    }

    var subEventIDs []uuid.UUID
    if in.SubEventID != nil {
        subEventIDs = []uuid.UUID{*in.SubEventID}
    }

    return a.upsert(ctx, orgID, eventID, peopleUpsertRequest{
        Email: in.Email, Mobile: in.Mobile, Name: in.Name,
        ImageURL: in.ImageURL, Age: in.Age, Gender: in.Gender,
        SubEventIDs: subEventIDs,
    }, !parent.IsDraft())
}

func (a *App) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (PersonResponse, error) {
    person, err := a.GetPerson(ctx, orgID, eventID, id)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrPeopleNotFound) {
            return PersonResponse{}, utils.ErrNotFound("Person not found.", err)
        }
        return PersonResponse{}, utils.ErrInternal("Failed to fetch person.", err)
    }
    return *person, nil
}

func (a *App) UpdateCardSentAt(ctx context.Context, id uuid.UUID) error {
    if err := a.MarkCardSent(ctx, id); err != nil {
        return utils.ErrInternal("Failed to mark card as sent.", err)
    }
    return nil
}

func (a *App) List(ctx context.Context, orgID, eventID uuid.UUID, filter PeopleListFilter) ([]PersonResponse, error) {
    searchStr := ""
    if filter.Search != nil {
        searchStr = *filter.Search
    }
    rows, err := a.ListPeople(ctx, orgID, eventID, filter.SubEventID, searchStr)
    if err != nil {
        return nil, utils.ErrInternal("Failed to list people.", err)
    }
    return rows, nil
}

func (a *App) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdatePersonRequest) (PersonResponse, error) {
    if err := a.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
        return PersonResponse{}, err
    }

    fields := PeopleFields{
        Name:     req.Name,
        ImageURL: strPtrOrNil(req.ImageURL),
        Age:      req.Age,
        Gender:   strPtrOrNil(req.Gender),
    }

    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        if _, err := a.UpdatePerson(ctx, tx, orgID, eventID, id, fields); err != nil {
            return err
        }
        if err := a.DeleteSubEventLinks(ctx, tx, id); err != nil {
            return err
        }
        for _, subEventID := range req.SubEventIDs {
            if err := a.AddToSubEvent(ctx, tx, id, subEventID); err != nil {
                return err
            }
        }
        return nil
    })
    if txErr != nil {
        return PersonResponse{}, utils.ErrInternal("Failed to update person.", txErr)
    }

    return a.Get(ctx, orgID, eventID, id)
}

func (a *App) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
    deleted, err := a.DeletePerson(ctx, orgID, eventID, id)
    if err != nil {
        return utils.ErrInternal("Failed to delete person.", err)
    }
    if !deleted {
        return utils.ErrNotFound("Person not found.", generic.ErrPeopleNotFound)
    }
    return nil
}

type peopleUpsertRequest struct {
    Email       string
    Mobile      string
    Name        string
    ImageURL    string
    Age         *int16
    Gender      string
    SubEventIDs []uuid.UUID
}

func (a *App) upsert(ctx context.Context, orgID, eventID uuid.UUID, req peopleUpsertRequest, backfillJoinedAt bool) (PersonResponse, bool, error) {
    if err := a.validateSubEventIDs(ctx, orgID, eventID, req.SubEventIDs); err != nil {
        return PersonResponse{}, false, err
    }

    var joinedAt *time.Time
    if backfillJoinedAt {
        now := time.Now().UTC()
        joinedAt = &now
    }

    fields := PeopleFields{
        Name:     req.Name,
        ImageURL: strPtrOrNil(req.ImageURL),
        Age:      req.Age,
        Gender:   strPtrOrNil(req.Gender),
    }

    var personResult *UpsertPersonResult
    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        upserted, err := a.UpsertPerson(ctx, tx, orgID, eventID, req.Email, req.Mobile, fields, joinedAt)
        if err != nil {
            return err
        }
        personResult = upserted

        if len(req.SubEventIDs) > 0 {
            for _, subEventID := range req.SubEventIDs {
                if err := a.AddToSubEvent(ctx, tx, upserted.ID, subEventID); err != nil {
                    return err
                }
            }
        }
        return nil
    })
    if txErr != nil {
        return PersonResponse{}, false, utils.ErrInternal("Failed to save person record.", txErr)
    }

    resp, err := a.Get(ctx, orgID, eventID, personResult.ID)
    if err != nil {
        return PersonResponse{}, false, err
    }
    return resp, personResult.Inserted, nil
}

func (a *App) validateSubEventIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) error {
    if len(ids) == 0 {
        return nil
    }
    validMap, err := a.subevents.ValidateIDs(ctx, orgID, eventID, ids)
    if err != nil {
        return err
    }
    for _, id := range ids {
        if !validMap[id] {
            return utils.ErrValidation(map[string]string{"sub_event_ids": "One or more sub_event_ids do not belong to this event."})
        }
    }
    return nil
}

var peopleCSVColumns = []string{"name", "email", "mobile"}

func (a *App) ImportCSV(ctx context.Context, orgID, eventID uuid.UUID, file io.Reader, opts PeopleImportOptions) (PeopleImportSummary, error) {
    parent, err := a.events.GetContext(ctx, orgID, eventID)
    if err != nil {
        return PeopleImportSummary{}, err
    }
    if opts.SubEventID != nil {
        if err := a.validateSubEventIDs(ctx, orgID, eventID, []uuid.UUID{*opts.SubEventID}); err != nil {
            return PeopleImportSummary{}, err
        }
    }

    reader := csv.NewReader(file)
    reader.TrimLeadingSpace = true
    header, err := reader.Read()
    if err != nil {
        return PeopleImportSummary{}, utils.NewError(http.StatusBadRequest, "Could not read CSV header.", err)
    }
    columns := make(map[string]int, len(header))
    for i, col := range header {
        columns[strings.ToLower(strings.TrimSpace(col))] = i
    }
    for _, required := range peopleCSVColumns {
        if _, ok := columns[required]; !ok {
            return PeopleImportSummary{}, utils.NewError(http.StatusBadRequest, "Missing required column: "+required, nil)
        }
    }

    var subEventIDs []uuid.UUID
    if opts.SubEventID != nil {
        subEventIDs = []uuid.UUID{*opts.SubEventID}
    }

    summary := PeopleImportSummary{}
    rowNum := 1
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        rowNum++
        if err != nil {
            summary.Skipped++
            summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "Malformed row."})
            continue
        }

        field := func(col string) string {
            idx, ok := columns[col]
            if !ok || idx >= len(record) {
                return ""
            }
            return strings.TrimSpace(record[idx])
        }

        name := field("name")
        email := strings.ToLower(field("email"))
        mobile := field("mobile")
        imageURL := field("image_url")
        ageStr := field("age")
        gender := field("gender")

        if name == "" || email == "" || mobile == "" {
            summary.Skipped++
            summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "name, email, and mobile are all required."})
            continue
        }
        if _, err := mail.ParseAddress(email); err != nil {
            summary.Skipped++
            summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "Invalid email address: " + email})
            continue
        }

        var age *int16
        if ageStr != "" {
            val, err := strconv.ParseInt(ageStr, 10, 16)
            if err != nil || val < 0 || val > 150 {
                summary.Skipped++
                summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: "Invalid age (must be 0-150): " + ageStr})
                continue
            }
            cast := int16(val)
            age = &cast
        }

        _, inserted, err := a.upsert(ctx, orgID, eventID, peopleUpsertRequest{
            Email: email, Mobile: mobile, Name: name,
            ImageURL: imageURL, Age: age, Gender: gender,
            SubEventIDs: subEventIDs,
        }, !parent.IsDraft())
        if err != nil {
            summary.Skipped++
            summary.Errors = append(summary.Errors, PeopleImportRowError{Row: rowNum, Message: err.Error()})
            continue
        }
        if inserted {
            summary.Inserted++
        } else {
            summary.Updated++
        }
    }

    return summary, nil
}

func (a *App) Export(ctx context.Context, orgID, eventID uuid.UUID, filter PeopleListFilter, ids []uuid.UUID) ([]byte, error) {
    var peopleList []PersonResponse
    var err error
    if len(ids) > 0 {
        peopleList, err = a.ListPeopleByIDs(ctx, orgID, eventID, ids)
    } else {
        peopleList, err = a.List(ctx, orgID, eventID, filter)
    }
    if err != nil {
        return nil, err
    }

    var buf bytes.Buffer
    w := csv.NewWriter(&buf)
    _ = w.Write([]string{"name", "email", "mobile", "image_url", "age", "gender", "joined_at", "card_sent_at"})
    for _, p := range peopleList {
        ageStr := ""
        if p.Age != nil {
            ageStr = strconv.Itoa(int(*p.Age))
        }
        _ = w.Write([]string{
            p.Name,
            p.Email,
            p.Mobile,
            derefStr(p.ImageURL),
            ageStr,
            derefStr(p.Gender),
            derefStr(p.JoinedAt),
            derefStr(p.CardSentAt),
        })
    }
    w.Flush()
    if err := w.Error(); err != nil {
        return nil, utils.ErrInternal("Failed to generate CSV export.", err)
    }
    return buf.Bytes(), nil
}

func derefStr(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

func strPtrOrNil(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
