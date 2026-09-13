"use client";

import { notFound, useParams, useRouter } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import {
    useDeletePersonMutation,
    usePersonDetailQuery,
    useUpdatePersonMutation,
} from "@/query-hooks/people.api";
import { useSubEventsListQuery } from "@/query-hooks/subevents.api";

import { PersonForm } from "../../person-form";
import { CardActions } from "./card-actions";
import { DeletePersonButton } from "./delete-person-button";

export default function EditPersonPage() {
    const { id, personId } = useParams<{ id: string; personId: string }>();
    const router = useRouter();

    const { data: person, error } = usePersonDetailQuery(id, personId);
    const { data: event } = useEventDetailQuery(id);
    const { data: subEvents = [] } = useSubEventsListQuery(id);

    const updatePersonMutation = useUpdatePersonMutation(id, personId);
    const deletePersonMutation = useDeletePersonMutation(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!person || !event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    async function updateAction(
        values: {
            name: string;
            image_url: string;
            age?: number;
            gender: string;
            sub_event_ids: string[];
        }
    ): Promise<{ error: string } | undefined> {
        const res = await updatePersonMutation.execute(values);
        if (res !== null) {
            router.push(`/dashboard/events/${id}/people`);
            return undefined;
        }
        return { error: updatePersonMutation.error?.message || "Something went wrong." };
    }

    async function deleteAction(): Promise<{ error: string } | undefined> {
        const res = await deletePersonMutation.execute(personId);
        if (res !== null) {
            router.push(`/dashboard/events/${id}/people`);
            return undefined;
        }
        return { error: deletePersonMutation.error?.message || "Something went wrong." };
    }

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-center justify-between">
                <h1 className="font-heading text-xl font-medium">Edit {person.name}</h1>
                <DeletePersonButton action={deleteAction} />
            </div>
            <CardActions
                eventId={id}
                personId={personId}
                downloadHref={`/dashboard/events/${id}/people/${personId}/card`}
                eventPublished={event.status === "published"}
                cardSentAt={person.card_sent_at}
            />
            <PersonForm
                subEvents={subEvents}
                showIdentity={false}
                defaultValues={{
                    email: person.email,
                    mobile: person.mobile,
                    name: person.name,
                    image_url: person.image_url ?? "",
                    age: person.age?.toString() ?? "",
                    gender: person.gender ?? "",
                    sub_event_ids: person.sub_event_ids,
                }}
                action={updateAction}
                submitLabel="Save changes"
            />
        </div>
    );
}