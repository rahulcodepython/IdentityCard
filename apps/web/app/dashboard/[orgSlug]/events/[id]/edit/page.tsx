"use client";

import Link from "next/link";
import { notFound, useParams } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";

import { EditEventForm } from "./edit-event-form";

export default function EditEventPage() {
    const params = useParams<{ id: string; orgSlug: string }>();
    const id = params?.id || "";
    const orgSlug = params?.orgSlug || "";

    const { data: event, error } = useEventDetailQuery(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    if (event.status !== "draft") {
        return (
            <div className="flex flex-col gap-3">
                <h1 className="font-heading text-xl font-medium">{event.name}</h1>
                <p className="text-sm text-muted-foreground">
                    This event is published — its schedule can no longer be edited.
                </p>
                <Link
                    href={`/dashboard/${orgSlug}/events/${id}`}
                    className="text-sm underline underline-offset-4"
                >
                    Back to event
                </Link>
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-center justify-between">
                <h1 className="font-heading text-xl font-medium">Edit {event.name}</h1>
                <div className="flex items-center gap-3 text-sm">
                    <Link
                        href={`/dashboard/${orgSlug}/events/${id}/days-import`}
                        className="underline underline-offset-4"
                    >
                        Import CSV
                    </Link>
                    <a
                        href={`/dashboard/${orgSlug}/events/${id}/days/export`}
                        className="underline underline-offset-4"
                    >
                        Export days CSV
                    </a>
                </div>
            </div>
            <EditEventForm eventId={id} event={event} />
        </div>
    );
}