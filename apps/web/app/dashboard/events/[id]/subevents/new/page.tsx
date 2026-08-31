"use client";

import { notFound, useParams, useRouter } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { useCreateSubEventMutation } from "@/query-hooks/subevents.api";

import { SubEventForm } from "../sub-event-form";

export default function NewSubEventPage() {
    const { id } = useParams<{ id: string }>();
    const router = useRouter();

    const { data: event, error } = useEventDetailQuery(id);
    const createSubEventMutation = useCreateSubEventMutation(id);

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
            <p className="text-sm text-muted-foreground">
                Sub-events can only be added while the event is a draft.
            </p>
        );
    }

    async function action(input: {
        name: string;
        date: string;
        entry_time: string;
        exit_time: string;
    }): Promise<{ error: string } | undefined> {
        const res = await createSubEventMutation.execute(input);
        if (res !== null) {
            router.push(`/dashboard/events/${id}`);
            return undefined;
        }
        return { error: createSubEventMutation.error?.message || "Something went wrong." };
    }

    return (
        <div className="flex flex-col gap-6">
            <h1 className="font-heading text-xl font-medium">New sub-event</h1>
            <SubEventForm
                eventDays={event.days}
                action={action}
                submitLabel="Create sub-event"
            />
        </div>
    );
}