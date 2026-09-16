"use client";

import * as React from "react";
import { Calendar, ChevronLeft, ChevronRight, Search } from "lucide-react";
import { cn } from "cn";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import type { AttendeeAnalysisItem } from "@/schema/attendance.types";

export interface AvailableDateOption {
    date: string;
    start_time?: string | null;
    end_time?: string | null;
}

interface AnalysisAttendeesTableProps {
    data: AttendeeAnalysisItem[];
    total: number;
    page: number;
    limit: number;
    search: string;
    status: string;
    hasDateRange: boolean;
    selectedDate: string;
    availableDates: AvailableDateOption[];
    onSearchChange: (val: string) => void;
    onStatusChange: (val: string) => void;
    onSelectedDateChange: (val: string) => void;
    onPageChange: (newPage: number) => void;
    isLoading: boolean;
}

export function AnalysisAttendeesTable({
    data,
    total,
    page,
    limit,
    search,
    status,
    hasDateRange,
    selectedDate,
    availableDates,
    onSearchChange,
    onStatusChange,
    onSelectedDateChange,
    onPageChange,
    isLoading,
}: AnalysisAttendeesTableProps) {
    const totalPages = Math.max(1, Math.ceil(total / limit));

    const formatTime = (timeStr?: string | null) => {
        if (!timeStr) return "-";
        const d = new Date(timeStr);
        return isNaN(d.getTime()) ? timeStr : d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    };

    return (
        <Card className="shadow-xs border-border/80">
            <CardHeader className="pb-3">
                <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div>
                        <CardTitle className="text-sm font-semibold">Attendee Roster</CardTitle>
                        <CardDescription className="text-xs">
                            Select a session date to audit entries, departures, and timings
                        </CardDescription>
                    </div>

                    <div className="flex flex-wrap items-center gap-2">
                        {/* Session Date Selector */}
                        <div className="flex items-center gap-1.5">
                            <Calendar className="size-3.5 text-muted-foreground" />
                            <select
                                value={selectedDate}
                                disabled={!hasDateRange || availableDates.length === 0}
                                onChange={(e) => onSelectedDateChange(e.target.value)}
                                className="h-8 rounded-md border border-input bg-background px-2 text-xs text-foreground focus-visible:outline-hidden focus-visible:ring-1 focus-visible:ring-ring font-medium disabled:opacity-50"
                            >
                                <option value="">
                                    {
                                        !hasDateRange
                                            ? "Select date range first..."
                                            : availableDates.length === 0
                                            ? "No dates in range"
                                            : "Select session date..."
                                    }
                                </option>
                                {
                                    availableDates.map((opt) => <option key={opt.date} value={opt.date}>
                                        {opt.date} {opt.start_time && opt.end_time ? `(${opt.start_time} - ${opt.end_time})` : ""}
                                    </option>
                                    )
                                }
                            </select>
                        </div>

                        {/* Search Input */}
                        <div className="relative">
                            <Search className="absolute left-2.5 top-2.5 size-3.5 text-muted-foreground" />
                            <Input
                                placeholder="Search attendees..."
                                value={search}
                                disabled={!selectedDate}
                                onChange={(e) => onSearchChange(e.target.value)}
                                className="h-8 w-40 pl-8 text-xs disabled:opacity-50"
                            />
                        </div>

                        {/* Status Filter */}
                        <select
                            value={status}
                            disabled={!selectedDate}
                            onChange={(e) => onStatusChange(e.target.value)}
                            className="h-8 rounded-md border border-input bg-background px-2 text-xs text-foreground focus-visible:outline-hidden focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50"
                        >
                            <option value="all">All Attendees</option>
                            <option value="attended">Attended</option>
                            <option value="inside">Currently Inside</option>
                            <option value="not_attended">Not Appeared</option>
                        </select>
                    </div>
                </div>
            </CardHeader>

            <CardContent className="p-0">
                <TooltipProvider delay={100}>
                    <div className="rounded-md border-t">
                        <Table>
                            <TableHeader>
                                <TableRow className="text-xs bg-muted/40">
                                    <TableHead className="font-semibold">Attendee</TableHead>
                                    <TableHead className="font-semibold">Session Date</TableHead>
                                    <TableHead className="font-semibold">Time Range</TableHead>
                                    <TableHead className="font-semibold">Check-In</TableHead>
                                    <TableHead className="font-semibold">Check-Out</TableHead>
                                    <TableHead className="font-semibold">Punctuality</TableHead>
                                    <TableHead className="font-semibold">Terminal</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody className="text-xs">
                                {
                                    !hasDateRange ? <TableRow>
                                        <TableCell colSpan={7} className="h-32 text-center text-muted-foreground">
                                            Please select a date range above to view attendance records.
                                        </TableCell>
                                    </TableRow> : !selectedDate ? <TableRow>
                                        <TableCell colSpan={7} className="h-32 text-center text-muted-foreground">
                                            Please select a session date from the dropdown above to load attendees.
                                        </TableCell>
                                    </TableRow> : isLoading ? <TableRow>
                                        <TableCell colSpan={7} className="h-32 text-center text-muted-foreground">
                                            Loading attendees...
                                        </TableCell>
                                    </TableRow> : data.length === 0 ? <TableRow>
                                        <TableCell colSpan={7} className="h-32 text-center text-muted-foreground">
                                            No attendees found for this session date.
                                        </TableCell>
                                    </TableRow> : data.map((item) => <TableRow key={`${item.applicant_id}-${item.date || "none"}`}>
                                        <TableCell>
                                            <div className="flex items-center gap-2.5">
                                                {/* Indicator Dot with Tooltip */}
                                                <Tooltip>
                                                    <TooltipTrigger
                                                        type="button"
                                                        className="inline-flex size-4 items-center justify-center rounded-full hover:opacity-80 focus-visible:outline-hidden cursor-help shrink-0"
                                                        aria-label={
                                                            item.status === "inside"
                                                                ? "Currently Inside"
                                                                : item.status === "attended"
                                                                ? "Attended"
                                                                : "Not Appeared"
                                                        }
                                                    >
                                                        <span
                                                            className={cn(
                                                                "size-2.5 rounded-full",
                                                                item.status === "inside" && "bg-blue-500 shadow-[0_0_6px_rgba(59,130,246,0.6)]",
                                                                item.status === "attended" && "bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.6)]",
                                                                item.status === "not_attended" && "bg-amber-500 shadow-[0_0_6px_rgba(245,158,11,0.6)]",
                                                            )}
                                                        />
                                                    </TooltipTrigger>
                                                    <TooltipContent side="top">
                                                        {
                                                            item.status === "inside"
                                                                ? "Currently Inside"
                                                                : item.status === "attended"
                                                                ? "Attended"
                                                                : "Not Appeared"
                                                        }
                                                    </TooltipContent>
                                                </Tooltip>

                                                <div className="flex flex-col">
                                                    <span className="font-semibold text-foreground">{item.name}</span>
                                                    <span className="font-mono text-[11px] text-muted-foreground">{item.email}</span>
                                                </div>
                                            </div>
                                        </TableCell>

                                        <TableCell className="text-muted-foreground">
                                            {item.date || selectedDate || "-"}
                                        </TableCell>

                                        <TableCell className="text-muted-foreground font-mono text-[11px]">
                                            {
                                                item.start_time && item.end_time
                                                    ? `${item.start_time} - ${item.end_time}`
                                                    : item.start_time || "-"
                                            }
                                        </TableCell>

                                        <TableCell className="font-medium text-foreground">
                                            {formatTime(item.entered_at)}
                                        </TableCell>

                                        <TableCell className="text-muted-foreground">
                                            {formatTime(item.exited_at)}
                                        </TableCell>

                                        <TableCell>
                                            {
                                                item.is_early === null || item.is_early === undefined ? <span className="text-muted-foreground/50">-</span> : item.is_early ? <span className="text-emerald-600 font-medium">On-Time</span> : <span className="text-amber-600 font-medium">Late</span>
                                            }
                                        </TableCell>

                                        <TableCell className="text-muted-foreground truncate max-w-[120px]">
                                            {item.device_name || "-"}
                                        </TableCell>
                                    </TableRow>
                                    )
                                }
                            </TableBody>
                        </Table>
                    </div>
                </TooltipProvider>

                {/* Pagination Controls */}
                {
                    selectedDate && total > 0 && <div className="flex items-center justify-between px-4 py-3 border-t text-xs text-muted-foreground">
                        <div>
                            Showing <span className="font-medium text-foreground">{data.length}</span> of{" "}
                            <span className="font-medium text-foreground">{total}</span> records
                        </div>

                        <div className="flex items-center gap-2">
                            <span className="text-xs">
                                Page {page} of {totalPages}
                            </span>
                            <div className="flex items-center gap-1">
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="icon"
                                    className="size-7"
                                    disabled={page <= 1 || isLoading}
                                    onClick={() => onPageChange(page - 1)}
                                >
                                    <ChevronLeft className="size-3.5" />
                                    <span className="sr-only">Previous Page</span>
                                </Button>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="icon"
                                    className="size-7"
                                    disabled={page >= totalPages || isLoading}
                                    onClick={() => onPageChange(page + 1)}
                                >
                                    <ChevronRight className="size-3.5" />
                                    <span className="sr-only">Next Page</span>
                                </Button>
                            </div>
                        </div>
                    </div>
                }
            </CardContent>
        </Card>
    );
}
