"use client";

import * as React from "react";

import { useCurrentEvent } from "@/components/events/event-context";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";

export default function EventSettingsPage() {
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
            url: `/dashboard/events/${eventId}`,
        },
        {
            title: "Settings",
        },
    ]);

    return (
        <div className="flex flex-1 flex-col gap-4">
            <div>
                <h1 className="text-2xl font-bold tracking-tight text-foreground">Settings</h1>
                <p className="text-xs text-muted-foreground">
                    Configure settings and details for {event.name}.
                </p>
            </div>
        </div>
    );
}