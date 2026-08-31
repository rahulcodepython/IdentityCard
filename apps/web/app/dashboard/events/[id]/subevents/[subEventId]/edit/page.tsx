"use client";

import { notFound, useParams, useRouter } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import {
    useDeleteSubEventMutation,
    useSubEventsListQuery,
    useUpdateSubEventMutation,
} from "@/query-hooks/subevents.api";

import { SubEventForm } from "../../sub-event-form";
import { DeleteSubEventButton } from "./delete-sub-event-button";

export default function EditSubEventPage() {
    const { id, subEventId } = useParams<{ id: string; subEventId: string }>();
    const router = useRouter();

    const { data: event, error: eventError } = useEventDetailQuery(id);
    const { data: subEvents = [] } = useSubEventsListQuery(id);
    const updateSubEventMutation = useUpdateSubEventMutation(id, subEventId);
    const deleteSubEventMutation = useDeleteSubEventMutation(id);

    const subEvent = subEvents.find((s) => s.id === subEventId);

    if (eventError instanceof ApiError && eventError.status === 404) {
        notFound();
    }

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    if (!subEvent) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading sub-event…
            </div>
        );
    }

    if (event.status !== "draft") {
        return (
            <p className="text-sm text-muted-foreground">
                Sub-events can only be edited while the event is a draft.
            </p>
        );
    }

    async function updateAction(input: {
        name: string;
        date: string;
        entry_time: string;
        exit_time: string;
    }): Promise<{ error: string } | undefined> {
        const res = await updateSubEventMutation.execute(input);
        if (res !== null) {
            router.push(`/dashboard/events/${id}`);
            return undefined;
        }
        return { error: updateSubEventMutation.error?.message || "Something went wrong." };
    }

    async function deleteAction(): Promise<{ error: string } | undefined> {
        const res = await deleteSubEventMutation.execute(subEventId);
        if (res !== null) {
            router.push(`/dashboard/events/${id}`);
            return undefined;
        }
        return { error: deleteSubEventMutation.error?.message || "Something went wrong." };
    }

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-center justify-between">
                <h1 className="font-heading text-xl font-medium">
                    Edit {subEvent.name}
                </h1>
                <DeleteSubEventButton action={deleteAction} />
            </div>
            <SubEventForm
                eventDays={event.days}
                defaultValues={{
                    name: subEvent.name,
                    date: subEvent.date,
                    entry_time: subEvent.entry_time,
                    exit_time: subEvent.exit_time,
                }}
                action={updateAction}
                submitLabel="Save changes"
            />
        </div>
    );
}