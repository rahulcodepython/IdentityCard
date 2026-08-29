package analytics

import "github.com/google/uuid"

type StatusCounts struct {
	Early  int `json:"early"`
	OnTime int `json:"on_time"`
	Late   int `json:"late"`
}

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

type SummaryResponse struct {
	TotalPeople int               `json:"total_people"`
	Attended    int               `json:"attended"`
	Absent      int               `json:"absent"`
	SubEvents   []SubEventSummary `json:"sub_events"`
}

type DailyResponse struct {
	Days []DailyBreakdown `json:"days"`
}

type DailyTrendPoint struct {
	Date    string `json:"date"`
	Present int    `json:"present"`
	Absent  int    `json:"absent"`
}

type OverviewResponse struct {
	TotalEvents    int               `json:"total_events"`
	EventsByStatus map[string]int    `json:"events_by_status"`
	TotalPeople    int               `json:"total_people"`
	Attended       int               `json:"attended"`
	Absent         int               `json:"absent"`
	ActiveDevices  int               `json:"active_devices"`
	DailyTrend     []DailyTrendPoint `json:"daily_trend"`
}
