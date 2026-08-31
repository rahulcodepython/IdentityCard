"use client";

import Link from "next/link";
import { notFound, useParams, useRouter, useSearchParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { ApiError } from "@/react-query/client";
import { useAnalyticsDailyQuery, useAnalyticsSummaryQuery } from "@/query-hooks/analytics.api";
import { useAttendanceRosterQuery } from "@/query-hooks/attendance.api";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { useSubEventsListQuery } from "@/query-hooks/subevents.api";

import { type ChartSeries, GroupedBarChart } from "./grouped-bar-chart";
import { RosterTable } from "./roster-table";

const STATUS_SERIES: ChartSeries[] = [
    { key: "early", label: "Early", color: "#d97706" },
    { key: "on_time", label: "On time", color: "var(--primary)" },
    { key: "late", label: "Late", color: "var(--destructive)" },
];
const ATTENDANCE_SERIES: ChartSeries[] = [
    { key: "present", label: "Present", color: "var(--primary)" },
    { key: "absent", label: "Absent", color: "var(--muted-foreground)" },
];

export default function AnalyticsPage() {
    const { id } = useParams<{ id: string }>();
    const router = useRouter();
    const searchParams = useSearchParams();

    const subEventId = searchParams.get("sub_event_id") ?? undefined;
    const date = searchParams.get("date") ?? undefined;
    const attended = searchParams.get("attended") ?? undefined;
    const status = searchParams.get("status") ?? undefined;

    const { data: event, error } = useEventDetailQuery(id);
    const { data: summary } = useAnalyticsSummaryQuery(id);
    const { data: daily = [] } = useAnalyticsDailyQuery(id, subEventId);
    const { data: subEvents = [] } = useSubEventsListQuery(id);
    const { data: roster = [] } = useAttendanceRosterQuery(id, {
        subEventId,
        date,
        attended: attended === "true" ? true : attended === "false" ? false : undefined,
        status,
    });

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event || !summary) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    const dailyCategories = daily.map((d) => d.date);
    const attendanceData = daily.map((d) => ({
        present: d.present,
        absent: d.absent,
    }));
    const entryStatusData = daily.map((d) => ({
        early: d.entry_statuses.early,
        on_time: d.entry_statuses.on_time,
        late: d.entry_statuses.late,
    }));
    const exitStatusData = daily.map((d) => ({
        early: d.exit_statuses.early,
        on_time: d.exit_statuses.on_time,
        late: d.exit_statuses.late,
    }));

    const exportParams = new URLSearchParams();
    if (subEventId) exportParams.set("sub_event_id", subEventId);
    if (date) exportParams.set("date", date);
    if (attended) exportParams.set("attended", attended);
    if (status) exportParams.set("status", status);

    function applyFilters(next: {
        subEventId?: string;
        date?: string;
        attended?: string;
        status?: string;
    }) {
        const params = new URLSearchParams();
        if (next.subEventId) params.set("sub_event_id", next.subEventId);
        if (next.date) params.set("date", next.date);
        if (next.attended) params.set("attended", next.attended);
        if (next.status) params.set("status", next.status);
        const qs = params.toString();
        router.push(`/dashboard/events/${id}/analytics${qs ? `?${qs}` : ""}`);
    }

    return (
        <div className="flex flex-col gap-8">
            <div className="flex items-center justify-between">
                <h1 className="font-heading text-xl font-medium">
                    Analytics — {event.name}
                </h1>
                <Button
                    variant="outline"
                    render={<Link href={`/dashboard/events/${id}/people`} />}
                >
                    People
                </Button>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <Card>
                    <CardHeader>
                        <CardDescription>Registered</CardDescription>
                        <CardTitle className="text-2xl">{summary.total_people}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader>
                        <CardDescription>Attended</CardDescription>
                        <CardTitle className="text-2xl">{summary.attended}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader>
                        <CardDescription>Absent</CardDescription>
                        <CardTitle className="text-2xl">{summary.absent}</CardTitle>
                    </CardHeader>
                </Card>
            </div>

            {summary.sub_events.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle>By sub-event</CardTitle>
                        <CardDescription>
                            Attended vs. absent, per sub-event
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <GroupedBarChart
                            categories={summary.sub_events.map((se) => se.sub_event_name)}
                            series={ATTENDANCE_SERIES}
                            data={summary.sub_events.map((se) => ({
                                present: se.attended,
                                absent: se.absent,
                            }))}
                            filename="attendance-by-sub-event.png"
                        />
                    </CardContent>
                </Card>
            )}

            <form
                onSubmit={(e) => {
                    e.preventDefault();
                    const fd = new FormData(e.currentTarget);
                    applyFilters({
                        subEventId: fd.get("sub_event_id")?.toString() || undefined,
                        date: fd.get("date")?.toString() || undefined,
                        attended: fd.get("attended")?.toString() || undefined,
                        status: fd.get("status")?.toString() || undefined,
                    });
                }}
                className="flex flex-wrap items-end gap-2 rounded-lg border p-3"
            >
                {subEvents.length > 0 && (
                    <div className="flex flex-col gap-1.5">
                        <label
                            className="text-xs text-muted-foreground"
                            htmlFor="sub_event_id"
                        >
                            Sub-event
                        </label>
                        <select
                            id="sub_event_id"
                            name="sub_event_id"
                            defaultValue={subEventId ?? ""}
                            className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                        >
                            <option value="">Whole event</option>
                            {subEvents.map((se) => (
                                <option key={se.id} value={se.id}>
                                    {se.name}
                                </option>
                            ))}
                        </select>
                    </div>
                )}
                <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-muted-foreground" htmlFor="date">
                        Date
                    </label>
                    <select
                        id="date"
                        name="date"
                        defaultValue={date ?? ""}
                        className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                    >
                        <option value="">All dates</option>
                        {event.days && event.days.map((d) => (
                            <option key={d.date} value={d.date}>
                                {d.date}
                            </option>
                        ))}
                    </select>
                </div>
                <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-muted-foreground" htmlFor="attended">
                        Attended
                    </label>
                    <select
                        id="attended"
                        name="attended"
                        defaultValue={attended ?? ""}
                        className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                    >
                        <option value="">Anyone</option>
                        <option value="true">Present</option>
                        <option value="false">Absent</option>
                    </select>
                </div>
                <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-muted-foreground" htmlFor="status">
                        Status
                    </label>
                    <select
                        id="status"
                        name="status"
                        defaultValue={status ?? ""}
                        className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                    >
                        <option value="">Any</option>
                        <option value="early">Early</option>
                        <option value="on_time">On time</option>
                        <option value="late">Late</option>
                    </select>
                </div>
                <Button type="submit" variant="outline" size="sm">
                    Apply
                </Button>
            </form>

            <Card>
                <CardHeader>
                    <CardTitle>Daily attendance</CardTitle>
                    <CardDescription>Present vs. absent per day</CardDescription>
                </CardHeader>
                <CardContent>
                    <GroupedBarChart
                        categories={dailyCategories}
                        series={ATTENDANCE_SERIES}
                        data={attendanceData}
                        filename="daily-attendance.png"
                    />
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Entry timing</CardTitle>
                    <CardDescription>Early / on time / late, per day</CardDescription>
                </CardHeader>
                <CardContent>
                    <GroupedBarChart
                        categories={dailyCategories}
                        series={STATUS_SERIES}
                        data={entryStatusData}
                        filename="entry-timing.png"
                    />
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Exit timing</CardTitle>
                    <CardDescription>Early / on time / late, per day</CardDescription>
                </CardHeader>
                <CardContent>
                    <GroupedBarChart
                        categories={dailyCategories}
                        series={STATUS_SERIES}
                        data={exitStatusData}
                        filename="exit-timing.png"
                    />
                </CardContent>
            </Card>

            <RosterTable
                roster={roster}
                exportHref={`/dashboard/events/${id}/analytics/export?${exportParams.toString()}`}
            />
        </div>
    );
}