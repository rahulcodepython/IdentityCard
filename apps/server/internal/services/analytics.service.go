// analytics rolls AttendanceService.BuildRoster's per-person data up
// into the numbers the spec asks admins be able to see: overall and
// per-sub-event attended/absent counts, and a per-day breakdown for
// multi-day events. It computes everything from that one roster rather
// than querying attendance_records itself, so there's exactly one place
// that defines what "expected" and "attended" mean. Leaf, aggregation-only
// — no repository of its own, same as CardsService.
package services

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/utils"
)

// overviewTrendDays caps how far back the dashboard home page's daily
// trend goes — long-running orgs would otherwise render an
// ever-growing, unreadable chart (the same "no pagination on the daily
// view" gap the per-event analytics page has, see PROJECT_MEMORY.md).
const overviewTrendDays = 30

type AnalyticsService struct {
	subevents  *SubEventsService
	attendance *AttendanceService
	events     *EventsService
	devices    *DevicesService
}

func NewAnalyticsService(subEventsService *SubEventsService, attendanceService *AttendanceService, eventsService *EventsService, devicesService *DevicesService) *AnalyticsService {
	return &AnalyticsService{subevents: subEventsService, attendance: attendanceService, events: eventsService, devices: devicesService}
}

func (s *AnalyticsService) Summary(ctx context.Context, orgID, eventID uuid.UUID) (entities.SummaryResponse, error) {
	total, attended, err := s.attendedCounts(ctx, orgID, eventID, nil)
	if err != nil {
		return entities.SummaryResponse{}, err
	}

	subEventList, err := s.subevents.List(ctx, orgID, eventID)
	if err != nil {
		return entities.SummaryResponse{}, err
	}

	subSummaries := make([]entities.SubEventSummary, len(subEventList))
	for i, se := range subEventList {
		seID := se.ID
		subTotal, subAttended, err := s.attendedCounts(ctx, orgID, eventID, &seID)
		if err != nil {
			return entities.SummaryResponse{}, err
		}
		subSummaries[i] = entities.SubEventSummary{
			SubEventID: se.ID, SubEventName: se.Name,
			TotalPeople: subTotal, Attended: subAttended, Absent: subTotal - subAttended,
		}
	}

	return entities.SummaryResponse{TotalPeople: total, Attended: attended, Absent: total - attended, SubEvents: subSummaries}, nil
}

func (s *AnalyticsService) Daily(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID) (entities.DailyResponse, error) {
	roster, err := s.attendance.BuildRoster(ctx, orgID, eventID, RosterFilter{SubEventID: subEventID})
	if err != nil {
		return entities.DailyResponse{}, err
	}

	type agg struct {
		expected, present int
		entry, exit       entities.StatusCounts
	}
	byDate := make(map[string]*agg)

	for _, e := range roster {
		a, ok := byDate[e.Date]
		if !ok {
			a = &agg{}
			byDate[e.Date] = a
		}
		a.expected++
		if e.Attended {
			a.present++
		}
		if e.EntryStatus != nil {
			incStatus(&a.entry, *e.EntryStatus)
		}
		if e.ExitStatus != nil {
			incStatus(&a.exit, *e.ExitStatus)
		}
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	days := make([]entities.DailyBreakdown, len(dates))
	for i, date := range dates {
		a := byDate[date]
		days[i] = entities.DailyBreakdown{
			Date: date, Expected: a.expected, Present: a.present, Absent: a.expected - a.present,
			EntryStatuses: a.entry, ExitStatuses: a.exit,
		}
	}
	return entities.DailyResponse{Days: days}, nil
}

// attendedCounts returns (total permitted people, how many actually
// showed up at least once) — optionally scoped to one sub-event's roster.
func (s *AnalyticsService) attendedCounts(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID) (total, attended int, err error) {
	roster, err := s.attendance.BuildRoster(ctx, orgID, eventID, RosterFilter{SubEventID: subEventID})
	if err != nil {
		return 0, 0, utils.ErrInternal()
	}

	allPeople := make(map[uuid.UUID]bool)
	attendedPeople := make(map[uuid.UUID]bool)
	for _, e := range roster {
		allPeople[e.PersonID] = true
		if e.Attended {
			attendedPeople[e.PersonID] = true
		}
	}
	return len(allPeople), len(attendedPeople), nil
}

// Overview rolls every event up into the numbers the dashboard home page
// shows — built the same way Summary/Daily are, from BuildRoster per
// event, so "expected"/"attended" never drift from the single per-event
// definition.
func (s *AnalyticsService) Overview(ctx context.Context, orgID uuid.UUID) (entities.OverviewResponse, error) {
	eventList, err := s.events.List(ctx, orgID)
	if err != nil {
		return entities.OverviewResponse{}, utils.ErrInternal()
	}

	resp := entities.OverviewResponse{
		TotalEvents:    len(eventList),
		EventsByStatus: map[string]int{},
	}

	cutoff := time.Now().AddDate(0, 0, -overviewTrendDays)
	trendByDate := map[string]*entities.DailyTrendPoint{}

	for _, event := range eventList {
		resp.EventsByStatus[event.Status]++

		roster, err := s.attendance.BuildRoster(ctx, orgID, event.ID, RosterFilter{})
		if err != nil {
			continue // a single broken event's roster shouldn't 500 the whole dashboard
		}

		allPeople := map[uuid.UUID]bool{}
		attendedPeople := map[uuid.UUID]bool{}
		for _, e := range roster {
			allPeople[e.PersonID] = true
			if e.Attended {
				attendedPeople[e.PersonID] = true
			}

			day, parseErr := time.Parse("2006-01-02", e.Date)
			if parseErr != nil || day.Before(cutoff) {
				continue
			}
			point, ok := trendByDate[e.Date]
			if !ok {
				point = &entities.DailyTrendPoint{Date: e.Date}
				trendByDate[e.Date] = point
			}
			if e.Attended {
				point.Present++
			} else {
				point.Absent++
			}
		}

		resp.TotalPeople += len(allPeople)
		resp.Attended += len(attendedPeople)
	}
	resp.Absent = resp.TotalPeople - resp.Attended

	dates := make([]string, 0, len(trendByDate))
	for date := range trendByDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	resp.DailyTrend = make([]entities.DailyTrendPoint, len(dates))
	for i, date := range dates {
		resp.DailyTrend[i] = *trendByDate[date]
	}

	deviceList, err := s.devices.List(ctx, orgID)
	if err == nil {
		for _, d := range deviceList {
			if d.Status == "verified" {
				resp.ActiveDevices++
			}
		}
	}

	return resp, nil
}

func incStatus(counts *entities.StatusCounts, status string) {
	switch status {
	case "early":
		counts.Early++
	case "on_time":
		counts.OnTime++
	case "late":
		counts.Late++
	}
}
