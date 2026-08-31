"use client";

import { type FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectiveDaysEditor } from "@/components/schedule/selective-days-editor";
import type { DayEntry } from "@/components/schedule/types";
import { useUpdateEventMutation } from "@/query-hooks/events.api";
import type { EventDetail, UpdateEventInput } from "@/schema/events.types";

const EVENT_TYPE_LABEL: Record<EventDetail["event_type"], string> = {
    flash: "Flash (single day) — fixed after creation",
    standard: "Standard (multi-day) — fixed after creation",
    grouped: "Grouped (range with sub-events) — fixed after creation",
};

export function EditEventForm({
    eventId,
    event,
}: {
    eventId: string;
    event: EventDetail;
}) {
    const [serverError, setServerError] = useState<string | null>(null);
    const [fieldErrors, setFieldErrors] = useState<Record<string, string> | null>(null);

    const [name, setName] = useState(event.name);
    const [nameError, setNameError] = useState<string | null>(null);
    const [venue, setVenue] = useState(event.venue ?? "");
    const [organizerName, setOrganizerName] = useState(event.organizer_name ?? "");

    const [days, setDays] = useState<DayEntry[]>(event.days);
    const [rangeStart, setRangeStart] = useState(event.start_date);
    const [rangeEnd, setRangeEnd] = useState(event.end_date ?? event.start_date);

    const router = useRouter();
    const updateMutation = useUpdateEventMutation(eventId);

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        if (name.trim().length < 2) {
            setNameError("Enter an event name");
            return;
        }
        setNameError(null);
        setServerError(null);
        setFieldErrors(null);

        const input: UpdateEventInput = {
            name,
            venue,
            organizer_name: organizerName,
            days: event.event_type !== "grouped" ? days : [],
            range_start: event.event_type === "grouped" ? rangeStart : "",
            range_end: event.event_type === "grouped" ? rangeEnd : "",
        };

        const result = await updateMutation.execute(input);
        if (result) {
            router.push(`/dashboard/events/${eventId}`);
        } else if (updateMutation.error) {
            setServerError(updateMutation.error.message || "Failed to update event");
        }
    }

    return (
        <form onSubmit={onSubmit} className="flex max-w-3xl flex-col gap-4" noValidate>
            <div className="flex flex-col gap-1.5">
                <Label htmlFor="name">Name</Label>
                <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    aria-invalid={!!nameError}
                />
                {nameError && <p className="text-xs text-destructive">{nameError}</p>}
            </div>

            <div className="flex flex-col gap-1.5">
                <Label>Event type</Label>
                <p className="text-sm text-muted-foreground">
                    {EVENT_TYPE_LABEL[event.event_type]}
                </p>
            </div>

            <div className="flex flex-col gap-1.5">
                <Label htmlFor="venue">Venue (optional)</Label>
                <Input id="venue" value={venue} onChange={(e) => setVenue(e.target.value)} />
            </div>

            <div className="flex flex-col gap-1.5">
                <Label htmlFor="organizer_name">Organizer Name (optional)</Label>
                <Input id="organizer_name" value={organizerName} onChange={(e) => setOrganizerName(e.target.value)} />
            </div>

            {(event.event_type === "flash" || event.event_type === "standard") && (
                <SelectiveDaysEditor
                    value={days}
                    onChange={setDays}
                    singleDay={event.event_type === "flash"}
                />
            )}

            {event.event_type === "grouped" && (
                <div className="grid grid-cols-2 gap-4">
                    <div className="flex flex-col gap-1.5">
                        <Label htmlFor="range_start">Range Start Date</Label>
                        <Input
                            id="range_start"
                            type="date"
                            value={rangeStart}
                            onChange={(e) => setRangeStart(e.target.value)}
                        />
                    </div>
                    <div className="flex flex-col gap-1.5">
                        <Label htmlFor="range_end">Range End Date</Label>
                        <Input
                            id="range_end"
                            type="date"
                            value={rangeEnd}
                            onChange={(e) => setRangeEnd(e.target.value)}
                        />
                    </div>
                </div>
            )}

            {fieldErrors && Object.keys(fieldErrors).length > 0 && (
                <ul className="list-disc pl-4 text-xs text-destructive">
                    {Object.entries(fieldErrors).map(([field, message]) => (
                        <li key={field}>
                            {field}: {message}
                        </li>
                    ))}
                </ul>
            )}

            {serverError && <p className="text-sm text-destructive">{serverError}</p>}

            <Button type="submit" disabled={updateMutation.isPending} className="mt-2 self-start">
                {updateMutation.isPending ? "Saving…" : "Save changes"}
            </Button>
        </form>
    );
}
