"use client";

import { notFound, useParams, useRouter } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { useCreatePersonMutation } from "@/query-hooks/people.api";
import { useSubEventsListQuery } from "@/query-hooks/subevents.api";

import { PersonForm } from "../person-form";

export default function NewPersonPage() {
    const { id } = useParams<{ id: string }>();
    const router = useRouter();

    const { data: event, error } = useEventDetailQuery(id);
    const { data: subEvents = [] } = useSubEventsListQuery(id);
    const createPersonMutation = useCreatePersonMutation(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    async function action(values: {
        email: string;
        mobile: string;
        name: string;
        image_url: string;
        age?: number;
        gender: string;
        sub_event_ids: string[];
    }) {
        const res = await createPersonMutation.execute(values);
        if (res !== null) {
            router.push(`/dashboard/events/${id}/people`);
            return undefined;
        }
        return { error: createPersonMutation.error?.message || "Something went wrong." };
    }

    return (
        <div className="flex flex-col gap-6">
            <h1 className="font-heading text-xl font-medium">
                Add person — {event.name}
            </h1>
            <PersonForm
                subEvents={subEvents}
                showIdentity
                action={action}
                submitLabel="Add person"
            />
        </div>
    );
}