// Package analytics rolls attendance.Service.BuildRoster's per-person
// data up into the numbers the spec asks admins be able to see: overall
// and per-sub-event attended/absent counts, and a per-day breakdown for
// multi-day events. It computes everything from that one roster rather
// than querying attendance_records itself, so there's exactly one place
// that defines what "expected" and "attended" mean.
package analytics

import (
	"context"
	"sort"

	"github.com/google/uuid"

	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/attendance"
	"identitycard-server/internal/modules/subevents"
)

type Service struct {
	subevents  *subevents.Service
	attendance *attendance.Service
}

func NewService(subEventsService *subevents.Service, attendanceService *attendance.Service) *Service {
	return &Service{subevents: subEventsService, attendance: attendanceService}
}

func (s *Service) Summary(ctx context.Context, orgID, eventID uuid.UUID) (SummaryResponse, error) {
	total, attended, err := s.attendedCounts(ctx, orgID, eventID, nil)
	if err != nil {
		return SummaryResponse{}, err
	}

	subEventList, err := s.subevents.List(ctx, orgID, eventID)
	if err != nil {
		return SummaryResponse{}, err
	}

	subSummaries := make([]SubEventSummary, len(subEventList))
	for i, se := range subEventList {
		seID := se.ID
		subTotal, subAttended, err := s.attendedCounts(ctx, orgID, eventID, &seID)
		if err != nil {
			return SummaryResponse{}, err
		}
		subSummaries[i] = SubEventSummary{
			SubEventID: se.ID, SubEventName: se.Name,
			TotalPeople: subTotal, Attended: subAttended, Absent: subTotal - subAttended,
		}
	}

	return SummaryResponse{TotalPeople: total, Attended: attended, Absent: total - attended, SubEvents: subSummaries}, nil
}

func (s *Service) Daily(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID) (DailyResponse, error) {
	roster, err := s.attendance.BuildRoster(ctx, orgID, eventID, attendance.RosterFilter{SubEventID: subEventID})
	if err != nil {
		return DailyResponse{}, err
	}

	type agg struct {
		expected, present int
		entry, exit       StatusCounts
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

	days := make([]DailyBreakdown, len(dates))
	for i, date := range dates {
		a := byDate[date]
		days[i] = DailyBreakdown{
			Date: date, Expected: a.expected, Present: a.present, Absent: a.expected - a.present,
			EntryStatuses: a.entry, ExitStatuses: a.exit,
		}
	}
	return DailyResponse{Days: days}, nil
}

// attendedCounts returns (total permitted people, how many actually
// showed up at least once) — optionally scoped to one sub-event's roster.
func (s *Service) attendedCounts(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID) (total, attended int, err error) {
	roster, err := s.attendance.BuildRoster(ctx, orgID, eventID, attendance.RosterFilter{SubEventID: subEventID})
	if err != nil {
		return 0, 0, httpx.ErrInternal()
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

func incStatus(counts *StatusCounts, status string) {
	switch status {
	case "early":
		counts.Early++
	case "on_time":
		counts.OnTime++
	case "late":
		counts.Late++
	}
}
