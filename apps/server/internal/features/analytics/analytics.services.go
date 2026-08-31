package analytics

import (
    "context"
    "sort"
    "time"

    "github.com/google/uuid"

    "identitycard-server/internal/features/attendance"
    "identitycard-server/internal/utils"
)

const overviewTrendDays = 30

func (a *App) Summary(ctx context.Context, orgID, eventID uuid.UUID) (SummaryResponse, error) {
    total, attended, err := a.attendedCounts(ctx, orgID, eventID, nil)
    if err != nil {
        return SummaryResponse{}, err
    }

    subEventList, err := a.subevents.List(ctx, orgID, eventID)
    if err != nil {
        return SummaryResponse{}, err
    }

    subSummaries := make([]SubEventSummary, len(subEventList))
    for i, se := range subEventList {
        seID := se.ID
        subTotal, subAttended, err := a.attendedCounts(ctx, orgID, eventID, &seID)
        if err != nil {
            return SummaryResponse{}, err
        }
        subSummaries[i] = SubEventSummary{
            SubEventID:  se.ID,
            SubEventName: se.Name,
            TotalPeople: subTotal,
            Attended:    subAttended,
            Absent:      subTotal - subAttended,
        }
    }

    return SummaryResponse{
        TotalPeople: total,
        Attended:    attended,
        Absent:      total - attended,
        SubEvents:   subSummaries,
    }, nil
}

func (a *App) Daily(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID) (DailyResponse, error) {
    roster, err := a.attendance.BuildRoster(ctx, orgID, eventID, attendance.RosterFilter{SubEventID: subEventID})
    if err != nil {
        return DailyResponse{}, err
    }

    type agg struct {
        expected, present int
        entry, exit       StatusCounts
    }
    byDate := make(map[string]*agg)

    for _, e := range roster {
        ag, ok := byDate[e.Date]
        if !ok {
            ag = &agg{}
            byDate[e.Date] = ag
        }
        ag.expected++
        if e.Attended {
            ag.present++
        }
        if e.EntryStatus != nil {
            incStatus(&ag.entry, *e.EntryStatus)
        }
        if e.ExitStatus != nil {
            incStatus(&ag.exit, *e.ExitStatus)
        }
    }

    dates := make([]string, 0, len(byDate))
    for date := range byDate {
        dates = append(dates, date)
    }
    sort.Strings(dates)

    days := make([]DailyBreakdown, len(dates))
    for i, date := range dates {
        ag := byDate[date]
        days[i] = DailyBreakdown{
            Date:          date,
            Expected:      ag.expected,
            Present:       ag.present,
            Absent:        ag.expected - ag.present,
            EntryStatuses: ag.entry,
            ExitStatuses:  ag.exit,
        }
    }
    return DailyResponse{Days: days}, nil
}

func (a *App) attendedCounts(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID) (total, attended int, err error) {
    roster, err := a.attendance.BuildRoster(ctx, orgID, eventID, attendance.RosterFilter{SubEventID: subEventID})
    if err != nil {
        return 0, 0, utils.ErrInternal("Failed to build roster for attended counts.", err)
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

func (a *App) Overview(ctx context.Context, orgID uuid.UUID) (OverviewResponse, error) {
    eventList, err := a.events.List(ctx, orgID)
    if err != nil {
        return OverviewResponse{}, utils.ErrInternal("Failed to list events for overview.", err)
    }

    resp := OverviewResponse{
        TotalEvents:    len(eventList),
        EventsByStatus: map[string]int{},
    }

    cutoff := time.Now().AddDate(0, 0, -overviewTrendDays)
    trendByDate := map[string]*DailyTrendPoint{}

    for _, event := range eventList {
        resp.EventsByStatus[event.Status]++

        roster, err := a.attendance.BuildRoster(ctx, orgID, event.ID, attendance.RosterFilter{})
        if err != nil {
            continue
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
                point = &DailyTrendPoint{Date: e.Date}
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
    resp.DailyTrend = make([]DailyTrendPoint, len(dates))
    for i, date := range dates {
        resp.DailyTrend[i] = *trendByDate[date]
    }

    deviceList, err := a.devices.List(ctx, orgID)
    if err == nil {
        for _, d := range deviceList {
            if d.Status == "verified" {
                resp.ActiveDevices++
            }
        }
    }

    return resp, nil
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
