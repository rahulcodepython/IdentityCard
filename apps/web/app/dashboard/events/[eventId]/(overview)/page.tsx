"use client";

import * as React from "react";
import {
    BarChart3,
    Calendar,
    CalendarDays,
    Clock,
    FileText,
    MapPin,
    Settings,
    UserCheck,
    Users,
} from "lucide-react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useCurrentEvent } from "@/components/events/event-context";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";

function formatDateRange(startDateStr?: string, endDateStr?: string): string {
    if (!startDateStr || !endDateStr) return "Oct 15, 2026 - Oct 18, 2026";
    try {
        const start = new Date(startDateStr);
        const end = new Date(endDateStr);
        const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric", year: "numeric" };
        return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}`;
    } catch {
        return `${startDateStr} - ${endDateStr}`;
    }
}

export default function EventOverviewPage() {
    const { event, eventId } = useCurrentEvent();

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
        },
    ]);

    const isEnded = React.useMemo(() => {
        const todayStr = new Date().toISOString().split("T")[0];
        return event.end_date < todayStr;
    }, [event.end_date]);

    const basePath = `/dashboard/events/${eventId}`;

    return (
        <div className="flex flex-1 flex-col gap-6">
            {/* Header section */}
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                        <h1 className="text-2xl font-bold tracking-tight text-foreground">
                            {event.name}
                        </h1>
                        <Badge variant={isEnded ? "secondary" : "default"} className="text-xs">
                            {isEnded ? "Ended" : "Active"}
                        </Badge>
                    </div>
                    <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1">
                            <Calendar className="size-3.5" />
                            {formatDateRange(event.start_date, event.end_date)}
                        </span>
                        <span className="flex items-center gap-1">
                            <MapPin className="size-3.5" />
                            {event.venue}
                        </span>
                        <span className="flex items-center gap-1">
                            <Users className="size-3.5" />
                            {event.organizer}
                        </span>
                    </div>
                </div>
            </div>

            {/* Quick Metrics */}
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-xs font-medium text-muted-foreground">
                            Total Attendees
                        </CardTitle>
                        <Users className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">1,248</div>
                        <p className="text-xs text-muted-foreground">+12% from last week</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-xs font-medium text-muted-foreground">
                            Sub Events
                        </CardTitle>
                        <CalendarDays className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">6 Sessions</div>
                        <p className="text-xs text-muted-foreground">Across 3 days</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-xs font-medium text-muted-foreground">
                            Registration Forms
                        </CardTitle>
                        <FileText className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">3 Active</div>
                        <p className="text-xs text-muted-foreground">General, VIP, Speaker</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-xs font-medium text-muted-foreground">
                            Checked-in Rate
                        </CardTitle>
                        <UserCheck className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">78.4%</div>
                        <p className="text-xs text-muted-foreground">978 verified check-ins</p>
                    </CardContent>
                </Card>
            </div>

            {/* Event Sections Quick Links */}
            <div className="flex flex-col gap-3">
                <h2 className="text-sm font-semibold text-foreground">Event Sections</h2>
                <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    <Card className="hover:bg-muted/40 transition-colors">
                        <Link href={`${basePath}/subevents`} className="block p-5">
                            <div className="flex items-center gap-3">
                                <div className="rounded-md bg-primary/10 p-2 text-primary">
                                    <CalendarDays className="size-5" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-foreground">Sub Events</h3>
                                    <p className="text-xs text-muted-foreground">
                                        Manage tracks, days, and session schedules
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>

                    <Card className="hover:bg-muted/40 transition-colors">
                        <Link href={`${basePath}/people`} className="block p-5">
                            <div className="flex items-center gap-3">
                                <div className="rounded-md bg-primary/10 p-2 text-primary">
                                    <Users className="size-5" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-foreground">People</h3>
                                    <p className="text-xs text-muted-foreground">
                                        Attendees, guests, speakers, and ID cards
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>

                    <Card className="hover:bg-muted/40 transition-colors">
                        <Link href={`${basePath}/forms`} className="block p-5">
                            <div className="flex items-center gap-3">
                                <div className="rounded-md bg-primary/10 p-2 text-primary">
                                    <FileText className="size-5" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-foreground">Forms</h3>
                                    <p className="text-xs text-muted-foreground">
                                        Custom registration forms and capacity
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>

                    <Card className="hover:bg-muted/40 transition-colors">
                        <Link href={`${basePath}/attendance`} className="block p-5">
                            <div className="flex items-center gap-3">
                                <div className="rounded-md bg-primary/10 p-2 text-primary">
                                    <UserCheck className="size-5" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-foreground">Attendance</h3>
                                    <p className="text-xs text-muted-foreground">
                                        Live check-in scanning and attendee roster
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>

                    <Card className="hover:bg-muted/40 transition-colors">
                        <Link href={`${basePath}/analytics`} className="block p-5">
                            <div className="flex items-center gap-3">
                                <div className="rounded-md bg-primary/10 p-2 text-primary">
                                    <BarChart3 className="size-5" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-foreground">Analytics</h3>
                                    <p className="text-xs text-muted-foreground">
                                        Check-in rates, attendance trends, and charts
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>

                    <Card className="hover:bg-muted/40 transition-colors">
                        <Link href={`${basePath}/settings`} className="block p-5">
                            <div className="flex items-center gap-3">
                                <div className="rounded-md bg-primary/10 p-2 text-primary">
                                    <Settings className="size-5" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-foreground">Settings</h3>
                                    <p className="text-xs text-muted-foreground">
                                        Event dates, branding, badges, and details
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>
                </div>
            </div>
        </div>
    );
}
