// Package repositories: attendance turns a verified QR scan into an
// entry or exit record. It's the first consumer of pkg/qrtoken.Parse, and
// reads events, people, subevents, and devices through their Services —
// never their repositories.
package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/utils/timeutil"
)

type AttendanceRepository struct {
	q *dbgen.Queries
}

func NewAttendanceRepository(q *dbgen.Queries) *AttendanceRepository {
	return &AttendanceRepository{q: q}
}

func (r *AttendanceRepository) Get(ctx context.Context, personID uuid.UUID, date string) (dbgen.AttendanceRecord, error) {
	d, err := timeutil.ParseDate(date)
	if err != nil {
		return dbgen.AttendanceRecord{}, err
	}
	return r.q.GetAttendanceRecord(ctx, dbgen.GetAttendanceRecordParams{PersonID: personID, Date: d})
}

func (r *AttendanceRepository) CreateEntry(ctx context.Context, orgID, eventID, personID uuid.UUID, date string, at time.Time, status string, deviceID uuid.UUID) (dbgen.AttendanceRecord, error) {
	d, err := timeutil.ParseDate(date)
	if err != nil {
		return dbgen.AttendanceRecord{}, err
	}
	return r.q.CreateAttendanceEntry(ctx, dbgen.CreateAttendanceEntryParams{
		OrganizationID: orgID,
		EventID:        eventID,
		PersonID:       personID,
		Date:           d,
		EntryAt:        pgtype.Timestamptz{Time: at, Valid: true},
		EntryStatus:    pgtype.Text{String: status, Valid: true},
		EntryDeviceID:  pgtype.UUID{Bytes: [16]byte(deviceID), Valid: true},
	})
}

func (r *AttendanceRepository) RecordExit(ctx context.Context, personID uuid.UUID, date string, at time.Time, status string, deviceID uuid.UUID) (dbgen.AttendanceRecord, error) {
	d, err := timeutil.ParseDate(date)
	if err != nil {
		return dbgen.AttendanceRecord{}, err
	}
	return r.q.RecordAttendanceExit(ctx, dbgen.RecordAttendanceExitParams{
		PersonID:     personID,
		Date:         d,
		ExitAt:       pgtype.Timestamptz{Time: at, Valid: true},
		ExitStatus:   pgtype.Text{String: status, Valid: true},
		ExitDeviceID: pgtype.UUID{Bytes: [16]byte(deviceID), Valid: true},
	})
}

func (r *AttendanceRepository) ListForEvent(ctx context.Context, orgID, eventID uuid.UUID) ([]dbgen.ListAttendanceForEventRow, error) {
	return r.q.ListAttendanceForEvent(ctx, dbgen.ListAttendanceForEventParams{EventID: eventID, OrganizationID: orgID})
}
