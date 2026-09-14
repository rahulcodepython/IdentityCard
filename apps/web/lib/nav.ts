"use client";

import { useMemo } from "react";
import { ArrowLeft, Calendar, LayoutDashboard, Settings } from "lucide-react";

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
        items: [{ title: "Events", url: "/dashboard/events", icon: Calendar }],
    },
];

function createEventSingleNavLayout(slugOrId: string, eventName: string): NavNode[] {
    const basePath = `/dashboard/events/${slugOrId}`;

    return [
        { type: "back", icon: ArrowLeft, title: "Back To Events", url: "/dashboard/events" },
        {
            title: eventName,
            type: "group",
            items: [
                { title: "Overview", url: basePath, icon: LayoutDashboard },
                { title: "Settings", url: `${basePath}/settings`, icon: Settings },
            ],
        },
    ];
}

export function useNavLayout(pathname: string): NavNode[] {
    // Cheap regex on a primitive — no need to memoize this itself.
    const rawId = EVENT_ID_REGEX.exec(pathname)?.[1] ?? "";
    const eventId = rawId && !RESERVED_EVENT_SEGMENTS.has(rawId) ? rawId : "";

    const { data: event } = useEventQuery(eventId);

    return useMemo(() => {
        if (eventId && event?.name) {
            return createEventSingleNavLayout(eventId, event.name);
        }
        return ROOT_NAV_LAYOUT;
    }, [eventId, event?.name]);
}