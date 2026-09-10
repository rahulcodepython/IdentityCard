package postgres

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the common interface implemented by both *pgxpool.Pool and pgx.Tx.
// Accepting DBTX allows all query functions to run standalone or within a transaction
// without duplicating functions or logic.
type DBTX interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Layer 1 Sentinels — driver-level domain sentinels
var (
    ErrNotFound     = errors.New("requested resource not found")
    ErrForbidden    = errors.New("access denied for entity")
    ErrInvalidState = errors.New("invalid state machine transition")
    ErrConflict     = errors.New("resource conflict or constraint violation")
    ErrInternalDB   = errors.New("unexpected database error")
)

// MapPgError maps PostgreSQL SQLSTATE error codes and pgx errors into generic domain sentinels.
func MapPgError(err error) error {
    if err == nil {
        return nil
    }
    if errors.Is(err, pgx.ErrNoRows) {
        return ErrNotFound
    }
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "02000":
            return fmt.Errorf("%w: %s", ErrNotFound, pgErr.Message)
        case "42501":
            return fmt.Errorf("%w: %s", ErrForbidden, pgErr.Message)
        case "23505":
            return fmt.Errorf("%w: %s", ErrConflict, pgErr.Message)
        case "P0001", "P0002":
            return fmt.Errorf("%w: %s", ErrInvalidState, pgErr.Message)
        case "57014":
            return fmt.Errorf("%w: %s", context.DeadlineExceeded, pgErr.Message)
        default:
            return fmt.Errorf("%w [SQLSTATE %s]: %s", ErrInternalDB, pgErr.Code, pgErr.Message)
        }
    }
    return err
}

// QueryJSON executes a query returning a single JSONB document and deserializes it into *T.
func QueryJSON[T any](ctx context.Context, db DBTX, sqlQuery string, args ...any) (*T, error) {
    var raw []byte
    err := db.QueryRow(ctx, sqlQuery, args...).Scan(&raw)
    if err != nil {
        return nil, MapPgError(err)
    }
    if len(raw) == 0 {
        return nil, ErrNotFound
    }
    return DecodeJSON[T](raw)
}

// QueryJSONTx is an alias for QueryJSON when running within an explicit transaction.
func QueryJSONTx[T any](ctx context.Context, tx pgx.Tx, sqlQuery string, args ...any) (*T, error) {
    return QueryJSON[T](ctx, tx, sqlQuery, args...)
}

// QueryJSONSlice executes a query returning a JSONB array document and deserializes it into []T.
// If the result is empty or null, it returns a non-nil empty slice []T{}.
func QueryJSONSlice[T any](ctx context.Context, db DBTX, sqlQuery string, args ...any) ([]T, error) {
    var raw []byte
    err := db.QueryRow(ctx, sqlQuery, args...).Scan(&raw)
    if err != nil {
        return nil, MapPgError(err)
    }
    if len(raw) == 0 {
        return []T{}, nil
    }
    return DecodeJSONSlice[T](raw)
}

// QueryJSONSliceTx is an alias for QueryJSONSlice when running within an explicit transaction.
func QueryJSONSliceTx[T any](ctx context.Context, tx pgx.Tx, sqlQuery string, args ...any) ([]T, error) {
    return QueryJSONSlice[T](ctx, tx, sqlQuery, args...)
}

// Exec executes an INSERT/UPDATE/DELETE statement with no result set.
func Exec(ctx context.Context, db DBTX, sqlQuery string, args ...any) error {
    _, err := db.Exec(ctx, sqlQuery, args...)
    return MapPgError(err)
}

// ExecTx is an alias for Exec when running within an explicit transaction.
func ExecTx(ctx context.Context, tx pgx.Tx, sqlQuery string, args ...any) error {
    return Exec(ctx, tx, sqlQuery, args...)
}

// DecodeJSON decodes raw JSONB bytes into *T.
func DecodeJSON[T any](raw []byte) (*T, error) {
    if len(raw) == 0 || string(raw) == "null" {
        return nil, nil
    }
    var dst T
    if err := json.Unmarshal(raw, &dst); err != nil {
        return nil, fmt.Errorf("decode json: %w", err)
    }
    return &dst, nil
}

// DecodeJSONSlice decodes raw JSONB bytes into []T, returning a non-nil empty slice on empty data.
func DecodeJSONSlice[T any](raw []byte) ([]T, error) {
    if len(raw) == 0 || string(raw) == "null" {
        return []T{}, nil
    }
    var dst []T
    if err := json.Unmarshal(raw, &dst); err != nil {
        return nil, fmt.Errorf("decode json slice: %w", err)
    }
    if dst == nil {
        dst = []T{}
    }
    return dst, nil
}

// Status-code queries map an integer status column in a single CTE row to domain errors.
type StatusErrorMap map[int]error

// QueryWithStatus executes a query returning (status_code, json_data) and checks status against errMap.
func QueryWithStatus[T any](ctx context.Context, db DBTX, sqlQuery string, errMap StatusErrorMap, args ...any) (*T, error) {
    var statusCode int
    var raw []byte
    err := db.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode, &raw)
    if err != nil {
        return nil, MapPgError(err)
    }
    if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
        return nil, domainErr
    }
    return DecodeJSON[T](raw)
}

// QuerySliceWithStatus executes a query returning (status_code, json_array) and checks status against errMap.
func QuerySliceWithStatus[T any](ctx context.Context, db DBTX, sqlQuery string, errMap StatusErrorMap, args ...any) ([]T, error) {
    var statusCode int
    var raw []byte
    err := db.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode, &raw)
    if err != nil {
        return nil, MapPgError(err)
    }
    if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
        return nil, domainErr
    }
    return DecodeJSONSlice[T](raw)
}

// QueryIDWithStatus executes a mutation/delete query returning (status_code, id) and checks status against errMap.
func QueryIDWithStatus(ctx context.Context, db DBTX, sqlQuery string, errMap StatusErrorMap, args ...any) (string, error) {
    var statusCode int
    var id string
    err := db.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode, &id)
    if err != nil {
        return "", MapPgError(err)
    }
    if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
        return "", domainErr
    }
    return id, nil
}

// QueryStatusOnly executes a query returning only (status_code) and checks status against errMap.
func QueryStatusOnly(ctx context.Context, db DBTX, sqlQuery string, errMap StatusErrorMap, args ...any) error {
    var statusCode int
    err := db.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode)
    if err != nil {
        return MapPgError(err)
    }
    if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
        return domainErr
    }
    return nil
}

// Condition and CheckConditions for manual validation
type Condition struct {
    Failed bool
    Err    error
}

func CheckConditions(conds ...Condition) error {
    for _, c := range conds {
        if c.Failed && c.Err != nil {
            return c.Err
        }
    }
    return nil
}
