"use client";

import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { EventDay } from "@/schema/events.types";

export function SubEventForm({
    eventDays,
    defaultValues,
    action,
    submitLabel,
}: {
    eventDays: EventDay[];
    defaultValues?: {
        name: string;
        date: string;
        entry_time: string;
        exit_time: string;
    };
    action: (input: {
        name: string;
        date: string;
        entry_time: string;
        exit_time: string;
    }) => Promise<{ error: string } | undefined>;
    submitLabel: string;
}) {
    const [serverError, setServerError] = useState<string | null>(null);
    const [isPending, setIsPending] = useState(false);

    const [name, setName] = useState(defaultValues?.name ?? "");
    const [nameError, setNameError] = useState<string | null>(null);
    const [date, setDate] = useState(defaultValues?.date ?? eventDays[0]?.date ?? "");
    const [entryTime, setEntryTime] = useState(
        defaultValues?.entry_time ?? eventDays[0]?.entry_time ?? "09:00"
    );
    const [exitTime, setExitTime] = useState(
        defaultValues?.exit_time ?? eventDays[0]?.exit_time ?? "17:00"
    );

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        if (name.trim().length < 2) {
            setNameError("Enter a sub-event name");
            return;
        }
        if (!date) {
            setServerError("Please select a date");
            return;
        }
        setNameError(null);
        setServerError(null);
        setIsPending(true);

        const result = await action({
            name,
            date,
            entry_time: entryTime,
            exit_time: exitTime,
        });
        setIsPending(false);
        if (result?.error) setServerError(result.error);
    }

    return (
        <form onSubmit={onSubmit} className="flex max-w-lg flex-col gap-4" noValidate>
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
                <Label htmlFor="date">Date</Label>
                {eventDays && eventDays.length > 0 ? (
                    <select
                        id="date"
                        value={date}
                        onChange={(e) => setDate(e.target.value)}
                        className="h-9 rounded-lg border border-input bg-transparent px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                    >
                        {eventDays.map((d) => (
                            <option key={d.date} value={d.date}>
                                {d.date}
                            </option>
                        ))}
                    </select>
                ) : (
                    <Input
                        id="date"
                        type="date"
                        value={date}
                        onChange={(e) => setDate(e.target.value)}
                    />
                )}
            </div>

            <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                    <Label htmlFor="entry_time">Entry Time</Label>
                    <Input
                        id="entry_time"
                        type="time"
                        value={entryTime}
                        onChange={(e) => setEntryTime(e.target.value)}
                    />
                </div>
                <div className="flex flex-col gap-1.5">
                    <Label htmlFor="exit_time">Exit Time</Label>
                    <Input
                        id="exit_time"
                        type="time"
                        value={exitTime}
                        onChange={(e) => setExitTime(e.target.value)}
                    />
                </div>
            </div>

            {serverError && <p className="text-sm text-destructive">{serverError}</p>}

            <Button type="submit" disabled={isPending} className="mt-2 self-start">
                {isPending ? "Saving…" : submitLabel}
            </Button>
        </form>
    );
}
