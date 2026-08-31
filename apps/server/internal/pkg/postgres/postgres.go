package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Layer 1 Sentinels — driver-level domain sentinels
var (
	ErrNotFound     = errors.New("requested resource not found")
	ErrForbidden    = errors.New("access denied for entity")
	ErrInvalidState = errors.New("invalid state machine transition")
	ErrConflict     = errors.New("resource conflict or constraint violation")
	ErrInternalDB   = errors.New("unexpected database error")
)

// MapPgError maps postgres/driver errors into generic domain sentinels.
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
		default:
			return fmt.Errorf("%w [SQLSTATE %s]: %s", ErrInternalDB, pgErr.Code, pgErr.Message)
		}
	}
	return err
}

// Single JSON document -> *T
func QueryJSON[T any](ctx context.Context, pool *pgxpool.Pool, sqlQuery string, args ...any) (*T, error) {
	var raw []byte
	err := pool.QueryRow(ctx, sqlQuery, args...).Scan(&raw)
	if err != nil {
		return nil, MapPgError(err)
	}
	if len(raw) == 0 {
		return nil, ErrNotFound
	}
	return DecodeJSON[T](raw)
}

// Single JSON document within a transaction -> *T
func QueryJSONTx[T any](ctx context.Context, tx pgx.Tx, sqlQuery string, args ...any) (*T, error) {
	var raw []byte
	err := tx.QueryRow(ctx, sqlQuery, args...).Scan(&raw)
	if err != nil {
		return nil, MapPgError(err)
	}
	if len(raw) == 0 {
		return nil, ErrNotFound
	}
	return DecodeJSON[T](raw)
}

// JSON array document -> []T (never nil — empty slice on empty/null)
func QueryJSONSlice[T any](ctx context.Context, pool *pgxpool.Pool, sqlQuery string, args ...any) ([]T, error) {
	var raw []byte
	err := pool.QueryRow(ctx, sqlQuery, args...).Scan(&raw)
	if err != nil {
		return nil, MapPgError(err)
	}
	if len(raw) == 0 {
		return []T{}, nil
	}
	return DecodeJSONSlice[T](raw)
}

// JSON array document within a transaction -> []T
func QueryJSONSliceTx[T any](ctx context.Context, tx pgx.Tx, sqlQuery string, args ...any) ([]T, error) {
	var raw []byte
	err := tx.QueryRow(ctx, sqlQuery, args...).Scan(&raw)
	if err != nil {
		return nil, MapPgError(err)
	}
	if len(raw) == 0 {
		return []T{}, nil
	}
	return DecodeJSONSlice[T](raw)
}

// Exec executes an INSERT/UPDATE/DELETE with no result set.
func Exec(ctx context.Context, pool *pgxpool.Pool, sqlQuery string, args ...any) error {
	_, err := pool.Exec(ctx, sqlQuery, args...)
	return MapPgError(err)
}

// ExecTx executes an INSERT/UPDATE/DELETE with no result set within a transaction.
func ExecTx(ctx context.Context, tx pgx.Tx, sqlQuery string, args ...any) error {
	_, err := tx.Exec(ctx, sqlQuery, args...)
	return MapPgError(err)
}

// DecodeJSON decodes raw JSONB bytes into *T.
func DecodeJSON[T any](raw []byte) (*T, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var dst T
	if err := json.Unmarshal(raw, &dst); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	return &dst, nil
}

// DecodeJSONSlice decodes raw JSONB bytes into []T.
func DecodeJSONSlice[T any](raw []byte) ([]T, error) {
	if len(raw) == 0 {
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

func QueryWithStatus[T any](ctx context.Context, pool *pgxpool.Pool, sqlQuery string, errMap StatusErrorMap, args ...any) (*T, error) {
	var statusCode int
	var raw []byte
	err := pool.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode, &raw)
	if err != nil {
		return nil, MapPgError(err)
	}
	if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
		return nil, domainErr
	}
	return DecodeJSON[T](raw)
}

func QuerySliceWithStatus[T any](ctx context.Context, pool *pgxpool.Pool, sqlQuery string, errMap StatusErrorMap, args ...any) ([]T, error) {
	var statusCode int
	var raw []byte
	err := pool.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode, &raw)
	if err != nil {
		return nil, MapPgError(err)
	}
	if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
		return nil, domainErr
	}
	return DecodeJSONSlice[T](raw)
}

func QueryIDWithStatus(ctx context.Context, pool *pgxpool.Pool, sqlQuery string, errMap StatusErrorMap, args ...any) (string, error) {
	var statusCode int
	var id string
	err := pool.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode, &id)
	if err != nil {
		return "", MapPgError(err)
	}
	if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
		return "", domainErr
	}
	return id, nil
}

func QueryStatusOnly(ctx context.Context, pool *pgxpool.Pool, sqlQuery string, errMap StatusErrorMap, args ...any) error {
	var statusCode int
	err := pool.QueryRow(ctx, sqlQuery, args...).Scan(&statusCode)
	if err != nil {
		return MapPgError(err)
	}
	if domainErr, ok := errMap[statusCode]; ok && domainErr != nil {
		return domainErr
	}
	return nil
}

// Condition and CheckConditions for manual Scan validation
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
