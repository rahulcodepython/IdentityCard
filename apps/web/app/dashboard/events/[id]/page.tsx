"use client";

import Link from "next/link";
import { notFound, useParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { ApiError } from "@/react-query/client";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { useSubEventsListQuery } from "@/query-hooks/subevents.api";
import type { EventDetail } from "@/schema/events.types";

import { DeleteButton } from "./delete-button";
import { PublishButton } from "./publish-button";

const EVENT_TYPE_LABEL: Record<EventDetail["event_type"], string> = {
    flash: "Flash Event",
    standard: "Standard Event",
    grouped: "Grouped Event",
};

export default function EventDetailPage() {
    const { id } = useParams<{ id: string }>();

    const { data: event, error } = useEventDetailQuery(id);
    const { data: subEvents = [] } = useSubEventsListQuery(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    const isDraft = event.status === "draft";

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-start justify-between">
                <div>
                    <h1 className="font-heading text-xl font-medium">{event.name}</h1>
                    <p className="text-sm text-muted-foreground">
                        {EVENT_TYPE_LABEL[event.event_type]} ·{" "}
                        {isDraft ? "Draft" : "Published"}
                        {event.venue ? ` · ${event.venue}` : ""}
                    </p>
                    <p className="text-xs text-muted-foreground">
                        {event.start_date} – {event.end_date ?? "open-ended"}
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        variant="outline"
                        render={<Link href={`/dashboard/events/${id}/people`} />}
                    >
                        People
                    </Button>
                    <Button
                        variant="outline"
                        render={<Link href={`/dashboard/events/${id}/forms`} />}
                    >
                        Forms
                    </Button>
                    <Button
                        variant="outline"
                        render={<Link href={`/dashboard/events/${id}/analytics`} />}
                    >
                        Analytics
                    </Button>
                    <a
                        href={`/dashboard/events/${id}/days/export`}
                        className="text-sm underline underline-offset-4"
                    >
                        Export days CSV
                    </a>
                    {isDraft && (
                        <Button
                            variant="outline"
                            render={<Link href={`/dashboard/events/${id}/edit`} />}
                        >
                            Edit
                        </Button>
                    )}
                    {isDraft && <DeleteButton eventId={id} />}
                    {isDraft && <PublishButton eventId={id} />}
                </div>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Schedule Days</CardTitle>
                    <CardDescription>
                        Entry/exit windows for each day of the event
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex max-h-96 flex-col divide-y overflow-y-auto">
                        {event.days && event.days.length > 0 ? (
                            event.days.map((day) => (
                                <div
                                    key={day.date}
                                    className="flex items-center justify-between py-2 text-sm"
                                >
                                    <span>{day.date}</span>
                                    <span className="text-muted-foreground">
                                        {day.entry_time} – {day.exit_time}
                                    </span>
                                </div>
                            ))
                        ) : (
                            <p className="text-sm text-muted-foreground py-2">
                                Date range: {event.start_date} to {event.end_date}
                            </p>
                        )}
                    </div>
                </CardContent>
            </Card>

            <div className="flex flex-col gap-3">
                <div className="flex items-center justify-between">
                    <h2 className="font-heading text-lg font-medium">Sub-events</h2>
                    {isDraft && (
                        <Button
                            variant="outline"
                            render={<Link href={`/dashboard/events/${id}/subevents/new`} />}
                        >
                            Add sub-event
                        </Button>
                    )}
                </div>

                {subEvents.length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                        No sub-events — everyone assigned to this event is treated as one group.
                    </p>
                ) : (
                    <div className="flex flex-col gap-3">
                        {subEvents.map((subEvent) => (
                            <Card key={subEvent.id}>
                                <CardHeader>
                                    <CardTitle>
                                        {isDraft ? (
                                            <Link
                                                href={`/dashboard/events/${id}/subevents/${subEvent.id}/edit`}
                                                className="hover:underline"
                                            >
                                                {subEvent.name}
                                            </Link>
                                        ) : (
                                            subEvent.name
                                        )}
                                    </CardTitle>
                                    <CardDescription>
                                        Date: {subEvent.date} ({subEvent.entry_time} – {subEvent.exit_time})
                                    </CardDescription>
                                </CardHeader>
                            </Card>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}