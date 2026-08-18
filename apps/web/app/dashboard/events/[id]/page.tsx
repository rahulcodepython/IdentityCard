import Link from "next/link"
import { notFound } from "next/navigation"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"
import { listSubEvents } from "@/lib/api/subevents"

import { DeleteButton } from "./delete-button"
import { PublishButton } from "./publish-button"

export default async function EventDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params

  let event
  try {
    event = await getEvent(id)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }
  const subEvents = await listSubEvents(id)
  const isDraft = event.status === "draft"

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="font-heading text-xl font-medium">{event.name}</h1>
          <p className="text-sm text-muted-foreground">
            {event.kind === "flash" ? "Flash" : "Established"} ·{" "}
            {isDraft ? "Draft" : "Published"}
            {event.venue ? ` · ${event.venue}` : ""}
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
          <CardTitle>Schedule</CardTitle>
          <CardDescription>Entry/exit window for each day</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col divide-y">
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
