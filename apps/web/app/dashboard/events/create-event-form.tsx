"use client";

import Link from "next/link";
import { type FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectiveDaysEditor } from "@/components/schedule/selective-days-editor";
import type { DayEntry } from "@/components/schedule/types";
import { useCreateEventMutation } from "@/query-hooks/events.api";
import type { CreateEventInput, EventType } from "@/schema/events.types";

const EVENT_TYPES: { value: EventType; label: string; description: string }[] = [
    {
        value: "flash",
        label: "Flash Event",
        description: "Single-day event with fixed date and entry/exit window",
    },
    {
        value: "standard",
        label: "Standard Event",
        description: "Multi-day scheduled event with customizable calendar days",
    },
    {
        value: "grouped",
        label: "Grouped Event",
        description: "Extended event spanning a date range with distinct sub-events",
    },
];

export function CreateEventForm({ allowFlash = true }: { allowFlash?: boolean }) {
    const [serverError, setServerError] = useState<string | null>(null);
    const [fieldErrors, setFieldErrors] = useState<Record<string, string> | null>(null);

    const [name, setName] = useState("");
    const [nameError, setNameError] = useState<string | null>(null);
    const [venue, setVenue] = useState("");
    const [organizerName, setOrganizerName] = useState("");
    const [eventType, setEventType] = useState<EventType>(allowFlash ? "flash" : "standard");

    const [days, setDays] = useState<DayEntry[]>([
        {
            date: new Date().toISOString().slice(0, 10),
            entry_time: "09:00",
            exit_time: "17:00",
        },
    ]);
    const [rangeStart, setRangeStart] = useState(new Date().toISOString().slice(0, 10));
    const [rangeEnd, setRangeEnd] = useState(new Date().toISOString().slice(0, 10));

    const router = useRouter();
    const createMutation = useCreateEventMutation();

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        if (name.trim().length < 2) {
            setNameError("Enter an event name");
            return;
        }
        setNameError(null);
        setServerError(null);
        setFieldErrors(null);

        const input: CreateEventInput = {
            name,
            venue,
            organizer_name: organizerName,
            event_type: eventType,
            days: eventType !== "grouped" ? days : [],
            range_start: eventType === "grouped" ? rangeStart : "",
            range_end: eventType === "grouped" ? rangeEnd : "",
        };

        const result = await createMutation.execute(input);
        if (result) {
            router.push(`/dashboard/events/${result.id}`);
        } else if (createMutation.error) {
            setServerError(createMutation.error.message || "Failed to create event");
        }
    }

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
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
                <Label htmlFor="venue">Venue (optional)</Label>
                <Input id="venue" value={venue} onChange={(e) => setVenue(e.target.value)} />
            </div>

            <div className="flex flex-col gap-1.5">
                <Label htmlFor="organizer_name">Organizer Name (optional)</Label>
                <Input id="organizer_name" value={organizerName} onChange={(e) => setOrganizerName(e.target.value)} />
            </div>

            <div className="flex flex-col gap-2">
                <Label>Event Type</Label>
                <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                    {EVENT_TYPES.map((t) => (
                        <button
                            key={t.value}
                            type="button"
                            onClick={() => setEventType(t.value)}
                            className={
                                "flex flex-col items-start gap-1 rounded-lg border p-3 text-left transition-colors " +
                                (eventType === t.value
                                    ? "border-primary bg-primary/5 ring-1 ring-primary"
                                    : "hover:bg-muted/50")
                            }
                        >
                            <span className="text-sm font-semibold text-foreground">{t.label}</span>
                            <span className="text-xs text-muted-foreground">{t.description}</span>
                        </button>
                    ))}
                </div>
            </div>

            {(eventType === "flash" || eventType === "standard") && (
                <SelectiveDaysEditor
                    value={days}
                    onChange={setDays}
                    singleDay={eventType === "flash"}
                />
            )}

            {eventType === "grouped" && (
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

            {serverError && (
                <p className="text-sm text-destructive">
                    {serverError}{" "}
                    <Link href="/dashboard/billing" className="underline underline-offset-4">
                        Manage Plan & Quota
                    </Link>
                </p>
            )}

            <Button type="submit" disabled={createMutation.isPending} className="mt-2 self-start">
                {createMutation.isPending ? "Creating…" : "Create event"}
            </Button>
        </form>
    );
}
