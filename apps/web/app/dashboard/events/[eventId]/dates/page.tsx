"use client";

import * as React from "react";
import { parseISO } from "date-fns";
import { useQueryClient } from "@tanstack/react-query";

import { useCurrentEvent } from "@/components/events/event-context";
import {
    CalendarActionBar,
    CustomTimeDialog,
    DefaultTimeBar,
    EventCalendar,
    FileUploadSection,
    SchemaPreview,
    StagedDatesPreviewDialog,
    type StagedDateItem,
} from "@/components/events/dates";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";
import { queryKeys } from "@/react-query/query-keys";
import {
    useBulkUpdateEventDatesMutation,
    useEventDatesQuery,
    useOverwrideEventDatesMutation,
} from "@/query-hooks/event-dates.api";
import type { EventDate, EventDateItemInput } from "@/schema/event-dates.types";

function toDateKey(d: Date): string {
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
}

function toMonthKey(d: Date): string {
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, "0");
    return `${year}-${month}`;
}

export default function EventDatesPage() {
    const { event, eventId } = useCurrentEvent();
    const queryClient = useQueryClient();

    useBreadcrumbs([
        {
            title: "Dashboard",
            url: "/dashboard",
        },
        {
            title: "Events",
            url: "/dashboard/events",
        },
        {
            title: event.name,
            url: `/dashboard/events/${eventId}`,
        },
        {
            title: "Dates & Schedule",
        },
    ]);

    // Active calendar viewing month
    const initialMonth = React.useMemo(() => {
        try {
            return event.start_date ? parseISO(event.start_date) : new Date();
        } catch {
            return new Date();
        }
    }, [event.start_date]);

    const [currentMonth, setCurrentMonth] = React.useState<Date>(initialMonth);
    const monthKey = toMonthKey(currentMonth);

    // Global default times applied to newly selected dates
    const [defaultStartTime, setDefaultStartTime] = React.useState("09:00");
    const [defaultEndTime, setDefaultEndTime] = React.useState("18:00");

    // Staging state for dates in the calendar
    const [stagedDates, setStagedDates] = React.useState<
        Record<string, { startTime: string; endTime: string; isCustom: boolean }>
    >({});

    // Track which months have been initialized from server to prevent background refetches from resurrecting deselected dates
    const initializedMonths = React.useRef<Set<string>>(new Set());

    // Dialog states
    const [customTimeDialogDate, setCustomTimeDialogDate] = React.useState<string | null>(null);
    const [isCustomTimeDialogOpen, setIsCustomTimeDialogOpen] = React.useState(false);
    const [isStagedPreviewOpen, setIsStagedPreviewOpen] = React.useState(false);

    // React query for current month's saved dates
    const { data: savedDatesData } = useEventDatesQuery(eventId, monthKey);
    const savedDates = React.useMemo(() => savedDatesData ?? [], [savedDatesData]);
    const bulkUpdateMutation = useBulkUpdateEventDatesMutation(eventId);
    const overwrideMutation = useOverwrideEventDatesMutation(eventId);

    // Saved dates map for this month
    const savedDatesMap = React.useMemo(() => {
        const map: Record<string, { startTime: string; endTime: string }> = {};
        for (const d of savedDates) {
            map[d.date] = {
                startTime: d.start_time,
                endTime: d.end_time,
            };
        }
        return map;
    }, [savedDates]);

    // Initialize month in stagedDates when first loaded from server
    React.useEffect(() => {
        if (!savedDatesData) return;

        if (!initializedMonths.current.has(monthKey)) {
            setStagedDates((prev) => {
                const next = { ...prev };
                for (const d of savedDatesData) {
                    next[d.date] = {
                        startTime: d.start_time,
                        endTime: d.end_time,
                        isCustom: d.start_time !== defaultStartTime || d.end_time !== defaultEndTime,
                    };
                }
                return next;
            });
            initializedMonths.current.add(monthKey);
        }
    }, [savedDatesData, monthKey, defaultStartTime, defaultEndTime]);

    // Overwride all dates (called by file upload section)
    const handleOverwrideDates = async (newDates: EventDateItemInput[]) => {
        await overwrideMutation.mutateAsync({ dates: newDates });
        const nextStaged: Record<string, { startTime: string; endTime: string; isCustom: boolean }> = {};
        for (const d of newDates) {
            nextStaged[d.date] = {
                startTime: d.start_time,
                endTime: d.end_time,
                isCustom: d.start_time !== defaultStartTime || d.end_time !== defaultEndTime,
            };
        }
        setStagedDates(nextStaged);

        const currentMonthNewDates: EventDate[] = newDates
            .filter((d) => d.date.startsWith(monthKey))
            .map((d) => ({
                id: crypto.randomUUID(),
                event_id: eventId,
                date: d.date,
                start_time: d.start_time,
                end_time: d.end_time,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
            }));

        queryClient.setQueryData(
            queryKeys.eventDates.byMonth(eventId, monthKey),
            currentMonthNewDates
        );

        initializedMonths.current.clear();
        initializedMonths.current.add(monthKey);
    };

    // Calculate staged changes for bulk operations
    const stagedChanges: StagedDateItem[] = React.useMemo(() => {
        const list: StagedDateItem[] = [];

        for (const [date, info] of Object.entries(stagedDates)) {
            const saved = savedDatesMap[date];
            if (!saved) {
                list.push({
                    date,
                    startTime: info.startTime,
                    endTime: info.endTime,
                    isCustom: info.isCustom,
                    changeType: "added",
                });
            } else if (saved.startTime !== info.startTime || saved.endTime !== info.endTime) {
                list.push({
                    date,
                    startTime: info.startTime,
                    endTime: info.endTime,
                    isCustom: info.isCustom,
                    changeType: "updated",
                });
            }
        }

        for (const [date, info] of Object.entries(savedDatesMap)) {
            if (!stagedDates[date]) {
                list.push({
                    date,
                    startTime: info.startTime,
                    endTime: info.endTime,
                    changeType: "removed",
                });
            }
        }

        return list.sort((a, b) => a.date.localeCompare(b.date));
    }, [stagedDates, savedDatesMap]);

    // Check if a date is within event range
    const isWithinEventRange = React.useCallback(
        (dateStr: string) => {
            return dateStr >= event.start_date && dateStr <= event.end_date;
        },
        [event.start_date, event.end_date]
    );

    // Toggle date selection
    const handleToggleDate = (date: Date) => {
        const dateStr = toDateKey(date);
        if (!isWithinEventRange(dateStr)) return;

        setStagedDates((prev) => {
            const next = { ...prev };
            if (next[dateStr]) {
                delete next[dateStr];
            } else {
                next[dateStr] = {
                    startTime: defaultStartTime,
                    endTime: defaultEndTime,
                    isCustom: false,
                };
            }
            return next;
        });
    };

    // Right-click date cell
    const handleContextMenu = (date: Date) => {
        const dateStr = toDateKey(date);
        if (!isWithinEventRange(dateStr)) return;

        setCustomTimeDialogDate(dateStr);
        setIsCustomTimeDialogOpen(true);
    };

    // Save custom time
    const handleSaveCustomTime = (dateStr: string, startTime: string, endTime: string) => {
        setStagedDates((prev) => ({
            ...prev,
            [dateStr]: {
                startTime,
                endTime,
                isCustom: startTime !== defaultStartTime || endTime !== defaultEndTime,
            },
        }));
    };

    // Reset date to default time
    const handleResetDateToDefault = (dateStr: string) => {
        setStagedDates((prev) => {
            if (!prev[dateStr]) return prev;
            return {
                ...prev,
                [dateStr]: {
                    startTime: defaultStartTime,
                    endTime: defaultEndTime,
                    isCustom: false,
                },
            };
        });
    };

    // Apply imported dates from file upload
    const handleApplyImportedDates = (imported: EventDateItemInput[]) => {
        setStagedDates((prev) => {
            const next = { ...prev };
            for (const item of imported) {
                if (isWithinEventRange(item.date)) {
                    next[item.date] = {
                        startTime: item.start_time,
                        endTime: item.end_time,
                        isCustom: item.start_time !== defaultStartTime || item.end_time !== defaultEndTime,
                    };
                }
            }
            return next;
        });
    };

    // Bulk Save all staged dates (atomic delete of deselected + upsert of remaining)
    const handleBulkSave = async () => {
        const datesToUpsert: EventDateItemInput[] = Object.entries(stagedDates).map(([d, val]) => ({
            date: d,
            start_time: val.startTime,
            end_time: val.endTime,
        }));

        const datesToDelete = Object.keys(savedDatesMap).filter((d) => !stagedDates[d]);

        await bulkUpdateMutation.mutateAsync({
            upsertDates: datesToUpsert,
            deleteDates: datesToDelete,
        });

        // Immediately update the query cache so UI reflects the exact new state without race conditions
        const updatedSavedDates: EventDate[] = datesToUpsert
            .filter((d) => d.date.startsWith(monthKey))
            .map((item) => {
                const existing = savedDates.find((s) => s.date === item.date);
                return {
                    id: existing?.id || crypto.randomUUID(),
                    event_id: eventId,
                    date: item.date,
                    start_time: item.start_time,
                    end_time: item.end_time,
                    created_at: existing?.created_at || new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                };
            });

        queryClient.setQueryData(
            queryKeys.eventDates.byMonth(eventId, monthKey),
            updatedSavedDates
        );

        // Ensure deselected dates are cleaned out of stagedDates
        setStagedDates((prev) => {
            const next = { ...prev };
            for (const d of datesToDelete) {
                delete next[d];
            }
            return next;
        });

        initializedMonths.current.add(monthKey);
    };

    return (
        <div className="flex flex-col gap-6">
            {/* Header section */}
            <div className="flex flex-col gap-1">
                <div className="flex items-center gap-2">
                    <h1 className="text-2xl font-bold tracking-tight text-foreground">
                        Event Dates & Schedule
                    </h1>
                    <Badge variant="outline" className="text-xs">
                        {event.start_date} to {event.end_date}
                    </Badge>
                </div>
                <p className="text-xs text-muted-foreground">
                    Select active event dates on the calendar, configure hours, or batch-import schedules via JSON and CSV.
                </p>
            </div>

            {/* Split Layout: Left Card (Calendar), Right Card matches Left Card height */}
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2 items-start">
                <Card className="flex flex-col">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base font-semibold">Calendar Schedule</CardTitle>
                        <CardDescription className="text-xs">
                            Click to toggle date selection. Right-click any cell to assign custom start and end times.
                        </CardDescription>
                    </CardHeader>

                    <CardContent className="flex flex-col gap-4">
                        <DefaultTimeBar
                            startTime={defaultStartTime}
                            endTime={defaultEndTime}
                            onStartTimeChange={setDefaultStartTime}
                            onEndTimeChange={setDefaultEndTime}
                        />

                        <div className="rounded-lg border bg-card p-3">
                            <EventCalendar
                                currentMonth={currentMonth}
                                onMonthChange={setCurrentMonth}
                                eventStartDate={event.start_date}
                                eventEndDate={event.end_date}
                                stagedDates={stagedDates}
                                onToggleDate={handleToggleDate}
                                onContextMenu={handleContextMenu}
                            />
                        </div>

                        <CalendarActionBar
                            selectedCount={Object.keys(stagedDates).length}
                            stagedCount={stagedChanges.length}
                            isPending={bulkUpdateMutation.isPending}
                            onPreviewStaged={() => setIsStagedPreviewOpen(true)}
                            onBulkSave={handleBulkSave}
                        />
                    </CardContent>
                </Card>

                {/* Right Column: Schema & Import */}
                <Card className="flex flex-col h-full">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base font-semibold">Schema & Import</CardTitle>
                        <CardDescription className="text-xs">
                            View schema specifications, inspect current exported data, or bulk-import dates from a file.
                        </CardDescription>
                    </CardHeader>

                    <CardContent className="flex flex-1 flex-col gap-4 justify-between">
                        <SchemaPreview currentDates={savedDates} />

                        <FileUploadSection
                            eventStartDate={event.start_date}
                            eventEndDate={event.end_date}
                            onApplyDates={handleApplyImportedDates}
                            onOverwrideDates={handleOverwrideDates}
                            isReplacing={overwrideMutation.isPending}
                        />
                    </CardContent>
                </Card>
            </div>

            {/* Custom Time Modal on Right-Click */}
            {
                customTimeDialogDate && <CustomTimeDialog
                    open={isCustomTimeDialogOpen}
                    onOpenChange={setIsCustomTimeDialogOpen}
                    dateStr={customTimeDialogDate}
                    initialStartTime={stagedDates[customTimeDialogDate]?.startTime || defaultStartTime}
                    initialEndTime={stagedDates[customTimeDialogDate]?.endTime || defaultEndTime}
                    onSave={handleSaveCustomTime}
                    onResetToDefault={handleResetDateToDefault}
                />
            }

            {/* Preview Staged Calendar Changes Modal */}
            <StagedDatesPreviewDialog
                open={isStagedPreviewOpen}
                onOpenChange={setIsStagedPreviewOpen}
                title="Preview Staged Calendar Changes"
                description="Review the dates to be added, modified, or deleted before saving to the database."
                items={stagedChanges}
                actionLabel="Save Changes"
                onAction={async () => {
                    await handleBulkSave();
                    setIsStagedPreviewOpen(false);
                }}
                isActionPending={bulkUpdateMutation.isPending}
            />
        </div>
    );
}
