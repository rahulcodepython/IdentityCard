"use client";

import * as React from "react";
import { format } from "date-fns";
import { ChevronLeft, ChevronRight, Clock } from "lucide-react";

import { cn } from "cn";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";

const WEEKDAY_NAMES = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

export interface EventCalendarProps {
    currentMonth: Date;
    onMonthChange: (newMonth: Date) => void;
    eventStartDate: string;
    eventEndDate: string;
    stagedDates: Record<string, { startTime: string; endTime: string; isCustom: boolean }>;
    onToggleDate: (date: Date) => void;
    onContextMenu: (date: Date) => void;
}

interface CalendarCell {
    date: Date;
    dateKey: string;
    dayNumber: number;
    isCurrentMonth: boolean;
    isDisabled: boolean;
    isSelected: boolean;
    isCustom: boolean;
}

function toDateKey(d: Date): string {
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
}

export function EventCalendar({
    currentMonth,
    onMonthChange,
    eventStartDate,
    eventEndDate,
    stagedDates,
    onToggleDate,
    onContextMenu,
}: EventCalendarProps) {
    const year = currentMonth.getFullYear();
    const monthIndex = currentMonth.getMonth();

    const handlePrevMonth = () => {
        onMonthChange(new Date(year, monthIndex - 1, 1));
    };

    const handleNextMonth = () => {
        onMonthChange(new Date(year, monthIndex + 1, 1));
    };

    // Compute calendar grid weeks
    const weeks = React.useMemo(() => {
        const firstDayOfWeek = new Date(year, monthIndex, 1).getDay();
        const daysInCurrentMonth = new Date(year, monthIndex + 1, 0).getDate();
        const daysInPrevMonth = new Date(year, monthIndex, 0).getDate();

        const grid: CalendarCell[][] = [];
        let currentWeek: CalendarCell[] = [];

        // Trailing days from previous month
        for (let i = firstDayOfWeek - 1; i >= 0; i--) {
            const dayNum = daysInPrevMonth - i;
            const d = new Date(year, monthIndex - 1, dayNum);
            const dateKey = toDateKey(d);
            currentWeek.push({
                date: d,
                dateKey,
                dayNumber: dayNum,
                isCurrentMonth: false,
                isDisabled: true,
                isSelected: Boolean(stagedDates[dateKey]),
                isCustom: Boolean(stagedDates[dateKey]?.isCustom),
            });
        }

        // Days in current month
        for (let dayNum = 1; dayNum <= daysInCurrentMonth; dayNum++) {
            const d = new Date(year, monthIndex, dayNum);
            const dateKey = toDateKey(d);
            const inRange = dateKey >= eventStartDate && dateKey <= eventEndDate;
            currentWeek.push({
                date: d,
                dateKey,
                dayNumber: dayNum,
                isCurrentMonth: true,
                isDisabled: !inRange,
                isSelected: Boolean(stagedDates[dateKey]),
                isCustom: Boolean(stagedDates[dateKey]?.isCustom),
            });

            if (currentWeek.length === 7) {
                grid.push(currentWeek);
                currentWeek = [];
            }
        }

        // Leading days for next month
        if (currentWeek.length > 0) {
            let nextDayNum = 1;
            while (currentWeek.length < 7) {
                const d = new Date(year, monthIndex + 1, nextDayNum);
                const dateKey = toDateKey(d);
                currentWeek.push({
                    date: d,
                    dateKey,
                    dayNumber: nextDayNum,
                    isCurrentMonth: false,
                    isDisabled: true,
                    isSelected: Boolean(stagedDates[dateKey]),
                    isCustom: Boolean(stagedDates[dateKey]?.isCustom),
                });
                nextDayNum++;
            }
            grid.push(currentWeek);
        }

        return grid;
    }, [year, monthIndex, eventStartDate, eventEndDate, stagedDates]);

    return (
        <div className="flex w-full flex-col gap-4">
            {/* Month Navigator Toolbar */}
            <div className="flex w-full items-center justify-between px-2 py-1">
                <Button
                    type="button"
                    variant="outline"
                    onClick={handlePrevMonth}
                    title="Previous month"
                    className="gap-1.5 text-xs font-medium"
                >
                    <ChevronLeft className="size-4" />
                    <span>Previous</span>
                </Button>
                <span className="text-base sm:text-lg font-bold tracking-tight text-foreground">
                    {format(currentMonth, "MMMM yyyy")}
                </span>
                <Button
                    type="button"
                    variant="outline"
                    onClick={handleNextMonth}
                    title="Next month"
                    className="gap-1.5 text-xs font-medium"
                >
                    <span>Next</span>
                    <ChevronRight className="size-4" />
                </Button>
            </div>

            {/* Much wider & larger Calendar Table */}
            <div className="w-full">
                <table className="w-full table-fixed border-collapse">
                    <thead>
                        <tr>
                            {
                                WEEKDAY_NAMES.map((name) => <th
                                    key={name}
                                    className="py-2 text-center text-xs sm:text-sm font-semibold text-muted-foreground select-none"
                                >
                                    {name}
                                </th>)
                            }
                        </tr>
                    </thead>
                    <tbody>
                        {
                            weeks.map((week, wIdx) => <tr key={wIdx}>
                                {
                                    week.map((cell) => <td key={cell.dateKey} className="p-1 sm:p-1.5 text-center">
                                        <button
                                            type="button"
                                            disabled={cell.isDisabled || !cell.isCurrentMonth}
                                            onClick={() => onToggleDate(cell.date)}
                                            onContextMenu={(e) => {
                                                e.preventDefault();
                                                if (!cell.isDisabled && cell.isCurrentMonth) {
                                                    onContextMenu(cell.date);
                                                }
                                            }}
                                            className={cn(
                                                "relative w-full aspect-square min-h-13.5 sm:min-h-16 max-w-19 rounded-md font-bold text-base sm:text-lg transition-all flex flex-col items-center justify-center select-none mx-auto",
                                                !cell.isCurrentMonth && "text-muted-foreground/20 opacity-20 cursor-default",
                                                cell.isCurrentMonth && cell.isDisabled && "text-muted-foreground/30 opacity-30 cursor-not-allowed",
                                                cell.isCurrentMonth && !cell.isDisabled && !cell.isSelected && "text-foreground hover:bg-muted/70 cursor-pointer",
                                                cell.isSelected && "bg-primary text-primary-foreground font-bold shadow-sm cursor-pointer"
                                            )}
                                        >
                                            <span>{cell.dayNumber}</span>
                                            {
                                                cell.isSelected && <span
                                                    className={cn(
                                                        "absolute bottom-1.5 size-2 rounded-md ring-1 ring-background",
                                                        cell.isCustom ? "bg-amber-400" : "bg-primary-foreground"
                                                    )}
                                                    title={cell.isCustom ? "Custom hours set" : "Scheduled"}
                                                />
                                            }
                                        </button>
                                    </td>)
                                }
                            </tr>)
                        }
                    </tbody>
                </table>
            </div>
        </div>
    );
}
