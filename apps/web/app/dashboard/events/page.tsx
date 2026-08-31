"use client";

import { useEventsListQuery } from "@/query-hooks/events.api";
import { useListBillingQuery } from "@/query-hooks/plans.api";

import { CreateEventDialog } from "./create-event-dialog";
import { EventsTable } from "./events-table";

export default function EventsPage() {
    const { data: events = [] } = useEventsListQuery();
    const { data: subs } = useListBillingQuery();

    const allowFlash =
        subs?.credits.some((c) => c.type === "flash" && !c.is_restricted) ?? true;

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-center justify-between">
                <h1 className="font-heading text-xl font-medium">Events</h1>
                <CreateEventDialog allowFlash={allowFlash} />
            </div>

            <EventsTable data={events} />
        </div>
    );
}