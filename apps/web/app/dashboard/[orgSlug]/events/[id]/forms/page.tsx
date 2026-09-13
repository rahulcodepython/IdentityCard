"use client";

import { notFound, useParams } from "next/navigation";

import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { useFormsListQuery } from "@/query-hooks/forms.api";
import { useSubEventsListQuery } from "@/query-hooks/subevents.api";

import { CreateFormForm } from "./create-form-form";
import { FormRow } from "./form-row";

export default function FormsPage() {
    const { id } = useParams<{ id: string }>();

    const { data: event, error } = useEventDetailQuery(id);
    const { data: forms = [] } = useFormsListQuery(id);
    const { data: subEvents = [] } = useSubEventsListQuery(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    const subEventNames = new Map(subEvents.map((se) => [se.id, se.name]));

    return (
        <div className="flex flex-col gap-6">
            <h1 className="font-heading text-xl font-medium">
                Public forms — {event.name}
            </h1>
            <p className="text-sm text-muted-foreground">
                Share a link below to let people register themselves, optionally capped
                at a fixed number of submissions.
            </p>

            <CreateFormForm eventId={id} subEvents={subEvents} />

            {forms.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                    No public sign-up links yet.
                </p>
            ) : (
                <div className="flex flex-col gap-3">
                    {forms.map((form) => (
                        <FormRow
                            key={form.id}
                            eventId={id}
                            form={form}
                            subEventName={
                                form.sub_event_id
                                    ? subEventNames.get(form.sub_event_id)
                                    : undefined
                            }
                        />
                    ))}
                </div>
            )}
        </div>
    );
}