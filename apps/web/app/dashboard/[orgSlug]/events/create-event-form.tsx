"use client";

import Link from "next/link";
import { type FormEvent, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectiveDaysEditor } from "@/components/schedule/selective-days-editor";
import type { DayEntry } from "@/components/schedule/types";
import { useCreateEventMutation } from "@/query-hooks/events.api";
import type { CreateEventInput } from "@/schema/events.types";
import useOrganization from "@/hooks/use-organization";

export function CreateEventForm() {
    const params = useParams<{ orgSlug: string }>();
    const orgSlug = params?.orgSlug || "";
    const { isOwner } = useOrganization(orgSlug);
    const [serverError, setServerError] = useState<string | null>(null);

    const todayStr = new Date().toISOString().slice(0, 10);
    const [name, setName] = useState("");
    const [nameError, setNameError] = useState<string | null>(null);
    const [venue, setVenue] = useState("");
    const [organizerName, setOrganizerName] = useState("");
    const [startDate, setStartDate] = useState(todayStr);
    const [endDate, setEndDate] = useState(todayStr);
    const [showDaySchedule, setShowDaySchedule] = useState(false);

    const [days, setDays] = useState<DayEntry[]>([
        {
            date: todayStr,
            entry_time: "09:00",
            exit_time: "17:00",
        },
    ]);

    const router = useRouter();
    const createMutation = useCreateEventMutation();

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        if (name.trim().length < 2) {
            setNameError("Enter an event name (min 2 characters)");
            return;
        }
        setNameError(null);
        setServerError(null);

        const input: CreateEventInput = {
            name,
            venue,
            organizer_name: organizerName,
            start_date: startDate,
            end_date: endDate,
            days: showDaySchedule ? days : [],
        };

        const result = await createMutation.execute(input);
        if (result) {
            router.push(`/dashboard/${orgSlug}/events/${result.id}`);
        } else if (createMutation.error) {
            setServerError(createMutation.error.message || "Failed to create event");
        }
    }

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
            <div className="flex flex-col gap-1.5">
                <Label htmlFor="name">Event Name *</Label>
                <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g. Annual Tech Symposium 2026"
                    aria-invalid={!!nameError}
                />
                {nameError && <p className="text-xs text-destructive">{nameError}</p>}
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                    <Label htmlFor="start_date">Start Date *</Label>
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
                    <Label htmlFor="end_date">End Date *</Label>
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
                <Input
                    id="venue"
                    value={venue}
                    onChange={(e) => setVenue(e.target.value)}
                    placeholder="e.g. Grand Auditorium, Block B"
                />
            </div>

            <div className="flex flex-col gap-1.5">
                <Label htmlFor="organizer_name">Organizer Name (optional)</Label>
                <Input
                    id="organizer_name"
                    value={organizerName}
                    onChange={(e) => setOrganizerName(e.target.value)}
                    placeholder="e.g. Department of Engineering"
                />
            </div>

            <div className="flex items-center justify-between border-t pt-4">
                <div>
                    <h4 className="text-sm font-medium">Custom Daily Timings</h4>
                    <p className="text-xs text-muted-foreground">
                        Optionally configure specific entry and exit times per date.
                    </p>
                </div>
                <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setShowDaySchedule(!showDaySchedule)}
                >
                    {showDaySchedule ? "Hide Schedule" : "Configure Days"}
                </Button>
            </div>

            {showDaySchedule && (
                <SelectiveDaysEditor
                    value={days}
                    onChange={setDays}
                    singleDay={startDate === endDate}
                />
            )}

            {serverError && (
                <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
                    <p className="font-medium">{serverError}</p>
                    {serverError.toLowerCase().includes("credit") && (
                        <p className="mt-1 text-xs">
                            {isOwner ? (
                                <Link href={`/dashboard/${orgSlug}/billing`} className="font-semibold underline underline-offset-4">
                                    Go to Billing to purchase event credits &rarr;
                                </Link>
                            ) : (
                                <span className="text-muted-foreground">
                                    Please contact an organization owner to purchase event credits.
                                </span>
                            )}
                        </p>
                    )}
                    {serverError.toLowerCase().includes("maintenance") && (
                        <p className="mt-1 text-xs">
                            {isOwner ? (
                                <Link href={`/dashboard/${orgSlug}/billing`} className="font-semibold underline underline-offset-4">
                                    Go to Billing to renew annual maintenance &rarr;
                                </Link>
                            ) : (
                                <span className="text-muted-foreground">
                                    Please contact an organization owner to renew annual maintenance.
                                </span>
                            )}
                        </p>
                    )}
                </div>
            )}

            <div className="flex items-center gap-3 pt-2">
                <Button type="submit" disabled={createMutation.isPending}>
                    {createMutation.isPending ? "Creating…" : "Create event (Uses 1 Credit)"}
                </Button>
                <span className="text-xs text-muted-foreground">
                    Deducts 1 Event Credit upon creation.
                </span>
            </div>
        </form>
    );
}
