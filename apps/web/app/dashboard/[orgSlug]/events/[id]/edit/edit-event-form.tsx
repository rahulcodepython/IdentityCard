"use client";

import { type FormEvent, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectiveDaysEditor } from "@/components/schedule/selective-days-editor";
import type { DayEntry } from "@/components/schedule/types";
import { useUpdateEventMutation } from "@/query-hooks/events.api";
import type { EventDetail, UpdateEventInput } from "@/schema/events.types";

export function EditEventForm({
    eventId,
    event,
}: {
    eventId: string;
    event: EventDetail;
}) {
    const params = useParams<{ orgSlug: string }>();
    const orgSlug = params?.orgSlug || "";
    const router = useRouter();

    const [serverError, setServerError] = useState<string | null>(null);
    const [fieldErrors, setFieldErrors] = useState<Record<string, string> | null>(null);

    const [name, setName] = useState(event.name);
    const [nameError, setNameError] = useState<string | null>(null);
    const [venue, setVenue] = useState(event.venue ?? "");
    const [organizerName, setOrganizerName] = useState(event.organizer_name ?? "");
    const [startDate, setStartDate] = useState(event.start_date ?? "");
    const [endDate, setEndDate] = useState(event.end_date ?? event.start_date ?? "");
    const [days, setDays] = useState<DayEntry[]>(event.days ?? []);

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
            name: name.trim(),
            venue: venue.trim(),
            organizer_name: organizerName.trim(),
            start_date: startDate,
            end_date: endDate,
            days,
        };

        const result = await updateMutation.execute(input);
        if (result) {
            router.push(`/dashboard/${orgSlug}/events/${eventId}`);
        } else if (updateMutation.error) {
            setServerError(updateMutation.error.message || "Failed to update event");
        }
    }

    return (
        <form onSubmit={onSubmit} className="flex max-w-3xl flex-col gap-4" noValidate>
            <div className="flex flex-col gap-1.5">
                <Label htmlFor="name">Event Name *</Label>
                <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    aria-invalid={!!nameError}
                />
                {nameError && <p className="text-xs text-destructive">{nameError}</p>}
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                    <Label htmlFor="start_date">Start Date</Label>
                    <Input
                        id="start_date"
                        type="date"
                        value={startDate}
                        onChange={(e) => {
                            setStartDate(e.target.value);
                            if (e.target.value > endDate) {
                                setEndDate(e.target.value);
                            }
                        }}
                    />
                </div>
                <div className="flex flex-col gap-1.5">
                    <Label htmlFor="end_date">End Date</Label>
                    <Input
                        id="end_date"
                        type="date"
                        value={endDate}
                        min={startDate}
                        onChange={(e) => setEndDate(e.target.value)}
                    />
                </div>
            </div>

            <div className="flex flex-col gap-1.5">
                <Label htmlFor="venue">Venue (optional)</Label>
                <Input id="venue" value={venue} onChange={(e) => setVenue(e.target.value)} />
            </div>

            <div className="flex flex-col gap-1.5">
                <Label htmlFor="organizer_name">Organizer Name (optional)</Label>
                <Input id="organizer_name" value={organizerName} onChange={(e) => setOrganizerName(e.target.value)} />
            </div>

            <div className="pt-2">
                <SelectiveDaysEditor
                    value={days}
                    onChange={setDays}
                />
            </div>

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
