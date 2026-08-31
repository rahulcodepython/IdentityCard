"use client";

import { notFound, useParams } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { useSubEventsListQuery } from "@/query-hooks/subevents.api";

import { ImportForm } from "./import-form";

export default function ImportPeoplePage() {
    const { id } = useParams<{ id: string }>();

    const { data: event, error } = useEventDetailQuery(id);
    const { data: subEvents = [] } = useSubEventsListQuery(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-6">
            <h1 className="font-heading text-xl font-medium">
                Import people — {event.name}
            </h1>
            <ImportForm eventId={id} subEvents={subEvents} />
        </div>
    );
}