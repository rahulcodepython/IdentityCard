package analytics

import "github.com/google/uuid"

type StatusCounts struct {
	Early  int `json:"early"`
	OnTime int `json:"on_time"`
	Late   int `json:"late"`
}

// DailyBreakdown is one event day's numbers — the data behind "a graph of
// performance ... filtered by dates" from the spec.
type DailyBreakdown struct {
	Date          string       `json:"date"`
	Expected      int          `json:"expected"`
	Present       int          `json:"present"`
	Absent        int          `json:"absent"`
	EntryStatuses StatusCounts `json:"entry_statuses"`
	ExitStatuses  StatusCounts `json:"exit_statuses"`
}

type SubEventSummary struct {
	SubEventID   uuid.UUID `json:"sub_event_id"`
	SubEventName string    `json:"sub_event_name"`
	TotalPeople  int       `json:"total_people"`
	Attended     int       `json:"attended"`
	Absent       int       `json:"absent"`
}

// SummaryResponse answers "how many people attended or were absent" at a
// glance, for the whole event and broken down per sub-event (empty if the
// event has none).
type SummaryResponse struct {
	TotalPeople int               `json:"total_people"`
	Attended    int               `json:"attended"`
	Absent      int               `json:"absent"`
	SubEvents   []SubEventSummary `json:"sub_events"`
}

type DailyResponse struct {
	Days []DailyBreakdown `json:"days"`
}

// DailyTrendPoint is one day of the org-wide (all events combined) trend
// shown on the dashboard home page — a coarser cousin of DailyBreakdown,
// which is always scoped to one event.
type DailyTrendPoint struct {
	Date    string `json:"date"`
	Present int    `json:"present"`
	Absent  int    `json:"absent"`
}

// OverviewResponse answers "how is the org doing overall" for the
// dashboard home page: every event, not just one.
type OverviewResponse struct {
	TotalEvents    int               `json:"total_events"`
	EventsByStatus map[string]int    `json:"events_by_status"`
	TotalPeople    int               `json:"total_people"`
	Attended       int               `json:"attended"`
	Absent         int               `json:"absent"`
	ActiveDevices  int               `json:"active_devices"`
	DailyTrend     []DailyTrendPoint `json:"daily_trend"`
}
