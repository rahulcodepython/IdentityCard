// Package timeutil converts between the wire format used by event/sub-event
// day schedules ("2006-01-02" dates, "15:04" clock times) and the pgtype
// values sqlc generates for Postgres DATE/TIME columns.
package timeutil

import (
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgtype"
)

const (
    DateLayout  = "2006-01-02"
    ClockLayout = "15:04"
)

func ParseDate(s string) (pgtype.Date, error) {
    t, err := time.Parse(DateLayout, s)
    if err != nil {
        return pgtype.Date{}, err
    }
    return pgtype.Date{Time: t, Valid: true}, nil
}

func FormatDate(d pgtype.Date) string {
    if !d.Valid {
        return ""
    }
    return d.Time.Format(DateLayout)
}

func FormatDateDirect(t time.Time) string {
    return t.Format(DateLayout)
}

func ParseClock(s string) (pgtype.Time, error) {
    t, err := time.Parse(ClockLayout, s)
    if err != nil {
        return pgtype.Time{}, err
    }
    micros := (t.Hour()*3600+t.Minute()*60+t.Second())*1_000_000 + t.Nanosecond()/1_000
    return pgtype.Time{Microseconds: int64(micros), Valid: true}, nil
}

func FormatClock(t pgtype.Time) string {
    if !t.Valid {
        return ""
    }
    totalSeconds := t.Microseconds / 1_000_000
    return fmt.Sprintf("%02d:%02d", totalSeconds/3600, (totalSeconds%3600)/60)
}
