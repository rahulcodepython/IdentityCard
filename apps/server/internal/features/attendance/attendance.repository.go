package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/utils/timeutil"
)

func (a *App) GetAttendanceRecord(ctx context.Context, personID uuid.UUID, date string) (dbgen.AttendanceRecord, error) {
	d, err := timeutil.ParseDate(date)
	if err != nil {
		return dbgen.AttendanceRecord{}, err
	}
	return a.queries.GetAttendanceRecord(ctx, dbgen.GetAttendanceRecordParams{PersonID: personID, Date: d})
}

func (a *App) CreateAttendanceEntry(ctx context.Context, orgID, eventID, personID uuid.UUID, date string, at time.Time, status string, deviceID uuid.UUID) (dbgen.AttendanceRecord, error) {
	d, err := timeutil.ParseDate(date)
	if err != nil {
		return dbgen.AttendanceRecord{}, err
	}
	return a.queries.CreateAttendanceEntry(ctx, dbgen.CreateAttendanceEntryParams{
		OrganizationID: orgID,
		EventID:        eventID,
		PersonID:       personID,
		Date:           d,
		EntryAt:        pgtype.Timestamptz{Time: at, Valid: true},
		EntryStatus:    pgtype.Text{String: status, Valid: true},
		EntryDeviceID:  pgtype.UUID{Bytes: [16]byte(deviceID), Valid: true},
	})
}

func (a *App) RecordAttendanceExit(ctx context.Context, personID uuid.UUID, date string, at time.Time, status string, deviceID uuid.UUID) (dbgen.AttendanceRecord, error) {
	d, err := timeutil.ParseDate(date)
	if err != nil {
		return dbgen.AttendanceRecord{}, err
	}
	return a.queries.RecordAttendanceExit(ctx, dbgen.RecordAttendanceExitParams{
		PersonID:     personID,
		Date:         d,
		ExitAt:       pgtype.Timestamptz{Time: at, Valid: true},
		ExitStatus:   pgtype.Text{String: status, Valid: true},
		ExitDeviceID: pgtype.UUID{Bytes: [16]byte(deviceID), Valid: true},
	})
}

func (a *App) ListAttendanceForEvent(ctx context.Context, orgID, eventID uuid.UUID) ([]dbgen.ListAttendanceForEventRow, error) {
	return a.queries.ListAttendanceForEvent(ctx, dbgen.ListAttendanceForEventParams{EventID: eventID, OrganizationID: orgID})
}
