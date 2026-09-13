"use client";

import { notFound, useParams } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";

import { ImportForm } from "./import-form";

export default function DaysImportPage() {
    const { id } = useParams<{ id: string }>();

    const { data: event, error } = useEventDetailQuery(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    const showDaysImport =
        event.status === "draft" && event.event_type === "standard";

    if (!showDaysImport) {
        return (
            <p className="text-sm text-muted-foreground">
                CSV import is only available for standard events in draft status.
            </p>
        );
    }

    return (
        <div className="flex flex-col gap-6">
            <h1 className="font-heading text-xl font-medium">
                Import Schedule Days — {event.name}
            </h1>
            <ImportForm eventId={id} />
        </div>
    );
}