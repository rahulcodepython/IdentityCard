"use client"

import Link from "next/link"
import { notFound, useParams } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { listSubEvents } from "@/lib/client-api/subevents"
import { queryKeys } from "@/react-query/query-keys"
import type { EventDetail } from "@/lib/validation/events"

import { AddExclusionForm } from "./add-exclusion-form"
import { DeleteButton } from "./delete-button"
import { PublishButton } from "./publish-button"

const SCHEDULE_MODE_LABEL: Record<EventDetail["schedule_mode"], string> = {
  flash: "Flash",
  fixed_range: "Fixed range",
  selective: "Selective dates",
  recurring: "Recurring",
}

export default function EventDetailPage() {
  const { id } = useParams<{ id: string }>()

  const { data: event, error } = useQuery({
    queryKey: queryKeys.event(id),
    queryFn: () => getEvent(id),
    retry: false,
  })

  const { data: subEvents = [] } = useQuery({
    queryKey: queryKeys.subEvents(id),
    queryFn: () => listSubEvents(id),
  })

  if (error instanceof ApiError && error.status === 404) notFound()

  if (!event) {
    return (
      <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
        Loading…
      </div>
    )
  }

  const isDraft = event.status === "draft"

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="font-heading text-xl font-medium">{event.name}</h1>
          <p className="text-sm text-muted-foreground">
            {SCHEDULE_MODE_LABEL[event.schedule_mode]} ·{" "}
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

      {event.schedule_mode === "recurring" && event.recurrence && (
        <Card>
          <CardHeader>
            <CardTitle>Recurrence</CardTitle>
            <CardDescription>
              Starts {event.recurrence.starts_on} ·{" "}
              {event.recurrence.ends_on ? `ends ${event.recurrence.ends_on}` : "open-ended"}
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <div className="flex flex-col divide-y">
              {event.recurrence.weekdays.map((w) => (
                <div key={w.weekday} className="flex items-center justify-between py-2 text-sm">
                  <span>{["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"][w.weekday]}</span>
                  <span className="text-muted-foreground">
                    {w.entry_time} – {w.exit_time}
                  </span>
                </div>
              ))}
            </div>
            <div className="flex flex-col gap-2 border-t pt-3">
              <p className="text-sm font-medium">Excluded dates</p>
              <p className="text-xs text-muted-foreground">
                {event.excluded_dates && event.excluded_dates.length > 0
                  ? event.excluded_dates.join(", ")
                  : "None"}
              </p>
              <AddExclusionForm eventId={id} />
              <Link
                href={`/dashboard/events/${id}/days-import`}
                className="text-xs underline underline-offset-4"
              >
                Import exclusion dates from CSV
              </Link>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Schedule</CardTitle>
          <CardDescription>
            Entry/exit window for each materialized day
            {event.schedule_mode === "recurring" && " (rolling window — extends automatically)"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex max-h-96 flex-col divide-y overflow-y-auto">
            {event.days.map((day) => (
              <div
                key={day.date}
                className="flex items-center justify-between py-2 text-sm"
              >
                <span>{day.date}</span>
                <span className="text-muted-foreground">
                  {day.entry_time} – {day.exit_time}
                </span>
              </div>
            ))}
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
            No sub-events — everyone assigned to this event is treated as one
            group.
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
                    {subEvent.days.length === 0
                      ? "Runs every day of the event"
                      : subEvent.days.map((d) => d.date).join(", ")}
                  </CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}