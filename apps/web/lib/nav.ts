"use client";

import { useMemo } from "react";
import { ArrowLeft, Calendar, CalendarDays, FileText, LayoutDashboard, Settings, Users } from "lucide-react";

import { useEventQuery } from "@/query-hooks/events.api";
import type { NavNode } from "@/schema/sidebar.types";

// Only need the id/slug segment right after /dashboard/events/
const EVENT_ID_REGEX = /^\/dashboard\/events\/([^/]+)/;
const RESERVED_EVENT_SEGMENTS = new Set(["create", "new"]);

// Static shape — build once at module scope instead of on every call.
const ROOT_NAV_LAYOUT: NavNode[] = [
    {
        title: "Dashboard",
        type: "group",
        items: [{ title: "Dashboard", url: "/dashboard", icon: LayoutDashboard }],
    },
    {
        title: "Management",
        type: "group",
        items: [
            { title: "Events", url: "/dashboard/events", icon: Calendar },
            { title: "Forms", url: "/dashboard/forms", icon: FileText },
        ],
    },
];

function createEventSingleNavLayout(eventId: string, eventName: string): NavNode[] {
    const basePath = `/dashboard/events/${eventId}`;

    return [
        { type: "back", icon: ArrowLeft, title: "Back To Events", url: "/dashboard/events" },
        {
            title: eventName.slice(0, 32) + "...",
            type: "group",
            items: [
                { title: "Overview", url: basePath, icon: LayoutDashboard },
                { title: "Dates", url: `${basePath}/dates`, icon: CalendarDays },
                { title: "Registration Form", url: `${basePath}/form`, icon: FileText },
                { title: "Applicants", url: `${basePath}/applicants`, icon: Users },
                { title: "Settings", url: `${basePath}/settings`, icon: Settings },
            ],
        },
    ];
}

export function useNavLayout(pathname: string): NavNode[] {
    const rawEventId = EVENT_ID_REGEX.exec(pathname)?.[1] ?? "";
    const eventId = rawEventId && !RESERVED_EVENT_SEGMENTS.has(rawEventId) ? rawEventId : "";

    const { data: event } = useEventQuery(eventId);

    return useMemo(() => {
        if (eventId) {
            return createEventSingleNavLayout(eventId, event?.name || "Event");
        }
        return ROOT_NAV_LAYOUT;
    }, [eventId, event?.name]);
}