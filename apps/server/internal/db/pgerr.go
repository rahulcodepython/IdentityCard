package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

// IsUniqueViolation reports whether err is a Postgres unique-constraint
// violation, optionally narrowed to a specific constraint name (pass ""
// to match any). Used to turn a raw DB error into a friendly
// httpx.ErrConflict instead of leaking a 500.
func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != uniqueViolationCode {
		return false
	}
	return constraint == "" || pgErr.ConstraintName == constraint
}
