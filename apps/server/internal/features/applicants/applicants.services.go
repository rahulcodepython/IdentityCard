package applicants

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
    "identitycard-server/internal/pkg/postgres"
)

var (
    ErrInvalidEventID     = errors.New("invalid event id")
    ErrInvalidName        = errors.New("name cannot be empty")
    ErrInvalidEmail       = errors.New("valid email address is required")
    ErrAlreadyRegistered  = errors.New("an applicant with this email is already registered for this event")
    ErrApplicantNotFound  = errors.New("applicant not found")
)

var safeIdentRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func (s *App) ListApplicantsService(ctx context.Context, eventID, search string, filters []ApplicantFilter, page, limit int) (*ListApplicantsResponse, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }

    form, err := s.GetAssignedFormRepository(ctx, eventID)
    if err != nil && !errors.Is(err, postgres.ErrNotFound) {
        return nil, err
    }

    fieldMap := make(map[string]FormFieldItem)
    if form != nil {
        for _, f := range form.Fields {
            fieldMap[f.Key] = f
        }
    }

    args := []any{eventID}
    whereClauses := []string{"ea.event_id = $1"}

    search = strings.TrimSpace(search)
    if search != "" {
        args = append(args, search)
        paramIdx := len(args)
        whereClauses = append(whereClauses, fmt.Sprintf("(a.name ILIKE '%%' || $%d || '%%' OR a.email ILIKE '%%' || $%d || '%%' OR a.id ILIKE '%%' || $%d || '%%' OR a.data::text ILIKE '%%' || $%d || '%%')", paramIdx, paramIdx, paramIdx, paramIdx))
    }

    for _, flt := range filters {
        key := strings.TrimSpace(flt.Field)
        op := strings.ToLower(strings.TrimSpace(flt.Op))
        val := strings.TrimSpace(flt.Value)
        if key == "" || val == "" {
            continue
        }

        // Standard direct columns
        if key == "name" {
            switch op {
            case "eq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.name = $%d", len(args)))
            case "neq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.name != $%d", len(args)))
            case "starts_with":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.name ILIKE $%d || '%%'", len(args)))
            default: // contains
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.name ILIKE '%%' || $%d || '%%'", len(args)))
            }
            continue
        }

        if key == "email" {
            switch op {
            case "eq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.email = $%d", len(args)))
            case "neq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.email != $%d", len(args)))
            default:
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.email ILIKE '%%' || $%d || '%%'", len(args)))
            }
            continue
        }

        if key == "user_id" {
            switch op {
            case "eq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.id = $%d", len(args)))
            default:
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("a.id ILIKE '%%' || $%d || '%%'", len(args)))
            }
            continue
        }

        // Check if key is a valid identifier and exists in the form schema
        if !safeIdentRegex.MatchString(key) {
            continue
        }

        fieldDef, ok := fieldMap[key]
        if !ok {
            continue
        }

        fieldType := strings.ToLower(fieldDef.Type)
        switch fieldType {
        case "number":
            // Validate that val is a valid number to prevent postgres numeric cast errors
            if _, err := strconv.ParseFloat(val, 64); err != nil {
                continue
            }
            var opSql string
            switch op {
            case "gt":
                opSql = ">"
            case "gte":
                opSql = ">="
            case "lt":
                opSql = "<"
            case "lte":
                opSql = "<="
            case "neq":
                opSql = "!="
            default:
                opSql = "="
            }
            args = append(args, val)
            whereClauses = append(whereClauses, fmt.Sprintf("((a.data->>'%s') ~ '^-?[0-9]+(\\.[0-9]+)?$' AND ((a.data->>'%s')::numeric) %s $%d::numeric)", key, key, opSql, len(args)))

        case "date":
            var opSql string
            switch op {
            case "gt":
                opSql = ">"
            case "gte":
                opSql = ">="
            case "lt":
                opSql = "<"
            case "lte":
                opSql = "<="
            case "neq":
                opSql = "!="
            default:
                opSql = "="
            }
            args = append(args, val)
            whereClauses = append(whereClauses, fmt.Sprintf("((a.data->>'%s') ~ '^\\d{4}-\\d{2}-\\d{2}' AND ((a.data->>'%s')::date) %s $%d::date)", key, key, opSql, len(args)))

        case "time", "month", "week":
            var opSql string
            switch op {
            case "gt":
                opSql = ">"
            case "gte":
                opSql = ">="
            case "lt":
                opSql = "<"
            case "lte":
                opSql = "<="
            case "neq":
                opSql = "!="
            default:
                opSql = "="
            }
            args = append(args, val)
            whereClauses = append(whereClauses, fmt.Sprintf("((a.data->>'%s') IS NOT NULL AND (a.data->>'%s') %s $%d)", key, key, opSql, len(args)))

        case "checkbox":
            if op == "contains" {
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("((jsonb_typeof(a.data->'%s') = 'array' AND a.data->'%s' ? $%d) OR (a.data->>'%s') ILIKE '%%' || $%d || '%%')", key, key, len(args), key, len(args)))
            } else if op == "neq" {
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(NOT (jsonb_typeof(a.data->'%s') = 'array' AND a.data->'%s' ? $%d) AND COALESCE(a.data->>'%s', '') != $%d)", key, key, len(args), key, len(args)))
            } else { // eq or default
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("((jsonb_typeof(a.data->'%s') = 'array' AND a.data->'%s' ? $%d) OR COALESCE(a.data->>'%s', '') = $%d)", key, key, len(args), key, len(args)))
            }

        case "switch":
            if op == "neq" {
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(a.data->>'%s', 'false') != $%d", key, len(args)))
            } else {
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(a.data->>'%s', 'false') = $%d", key, len(args)))
            }

        case "radio", "select":
            switch op {
            case "neq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(a.data->>'%s') != $%d", key, len(args)))
            case "contains":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(a.data->>'%s') ILIKE '%%' || $%d || '%%'", key, len(args)))
            default: // eq
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(a.data->>'%s') = $%d", key, len(args)))
            }

        default: // text, textarea, email, url, file, etc.
            switch op {
            case "eq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(a.data->>'%s') = $%d", key, len(args)))
            case "neq":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(a.data->>'%s') != $%d", key, len(args)))
            case "starts_with":
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(a.data->>'%s') ILIKE $%d || '%%'", key, len(args)))
            default: // contains
                args = append(args, val)
                whereClauses = append(whereClauses, fmt.Sprintf("(a.data->>'%s') ILIKE '%%' || $%d || '%%'", key, len(args)))
            }
        }
    }

    offset := (page - 1) * limit
    args = append(args, limit, offset, page, limit)
    limitIdx := len(args) - 3
    offsetIdx := len(args) - 2
    pageIdx := len(args) - 1
    limitIdx2 := len(args)

    whereClauseStr := strings.Join(whereClauses, " AND ")

    paginated, err := s.QueryApplicantsWithFiltersRepository(ctx, whereClauseStr, limitIdx, offsetIdx, pageIdx, limitIdx2, args...)
    if err != nil {
        return nil, err
    }

    return &ListApplicantsResponse{
        Data:  paginated.Data,
        Total: paginated.Total,
        Page:  paginated.Page,
        Limit: paginated.Limit,
        Form:  nil,
    }, nil
}

func (s *App) GetApplicantSchemaService(ctx context.Context, eventID string) (*FormSummary, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }

    form, err := s.GetAssignedFormRepository(ctx, eventID)
    if err != nil && !errors.Is(err, postgres.ErrNotFound) {
        return nil, err
    }

    return form, nil
}

func (s *App) CreateApplicantService(ctx context.Context, eventID string, req CreateApplicantRequest) (*ApplicantItem, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }

    trimmedName := strings.TrimSpace(req.Name)
    if trimmedName == "" {
        return nil, ErrInvalidName
    }

    trimmedEmail := strings.ToLower(strings.TrimSpace(req.Email))
    if trimmedEmail == "" || !strings.Contains(trimmedEmail, "@") {
        return nil, ErrInvalidEmail
    }

    if req.Data == nil {
        req.Data = make(map[string]interface{})
    }

    dataJSON, err := json.Marshal(req.Data)
    if err != nil {
        return nil, err
    }

    eventClean := strings.ReplaceAll(eventID, "-", "")
    if len(eventClean) > 8 {
        eventClean = eventClean[:8]
    }
    randomClean := strings.ReplaceAll(uuid.NewString(), "-", "")
    if len(randomClean) > 8 {
        randomClean = randomClean[:8]
    }
    datePart := time.Now().UTC().Format("20060102")
    userID := fmt.Sprintf("%s-%s-%s", eventClean, randomClean, datePart)

    created, err := s.CreateApplicantRepository(ctx, eventID, userID, trimmedName, trimmedEmail, dataJSON)
    if err != nil {
        if errors.Is(err, postgres.ErrConflict) {
            return nil, ErrAlreadyRegistered
        }
        return nil, err
    }
    if created == nil {
        return nil, ErrAlreadyRegistered
    }

    return created, nil
}

func (s *App) DeleteApplicantService(ctx context.Context, eventID, userID string) (*DeleteApplicantResponse, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }
    userID = strings.TrimSpace(userID)
    if userID == "" {
        return nil, ErrApplicantNotFound
    }

    if err := s.DeleteApplicantRepository(ctx, eventID, userID); err != nil {
        return nil, err
    }

    return &DeleteApplicantResponse{
        UserID:  userID,
        Message: "Applicant deleted successfully",
    }, nil
}

