"use client";

import * as React from "react";

import type { Event } from "@/schema/events.types";

interface EventContextValue {
    event: Event;
    eventId: string;
}

const EventContext = React.createContext<EventContextValue | null>(null);

export interface EventProviderProps {
    event: Event;
    eventId: string;
    children: React.ReactNode;
}

export function EventProvider({
    event,
    eventId,
    children,
}: EventProviderProps) {
    return (
        <EventContext.Provider value={{ event, eventId }}>
            {children}
        </EventContext.Provider>
    );
}

/**
 * Hook to access the current validated event data in any child page or component under [eventId].
 * Guaranteed to return non-null event data since the parent layout handles loading and 404 checks.
 */
export function useCurrentEvent(): EventContextValue {
    const context = React.useContext(EventContext);
    if (!context) {
        throw new Error(
            "useCurrentEvent must be used within an EventProvider (under dashboard/events/[eventId]/layout.tsx)"
        );
    }
    return context;
}
