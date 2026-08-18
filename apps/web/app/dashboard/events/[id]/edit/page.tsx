import Link from "next/link"
import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"

import { EditEventForm } from "./edit-event-form"

export default async function EditEventPage({
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

  if (event.status !== "draft") {
    return (
      <div className="flex flex-col gap-3">
        <h1 className="font-heading text-xl font-medium">{event.name}</h1>
        <p className="text-sm text-muted-foreground">
          This event is published — its schedule can no longer be edited.
        </p>
        <Link
          href={`/dashboard/events/${id}`}
          className="text-sm underline underline-offset-4"
        >
          Back to event
        </Link>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">Edit {event.name}</h1>
      <EditEventForm
        eventId={id}
        kind={event.kind}
        defaultValues={{
          name: event.name,
          venue: event.venue ?? "",
          days: event.days,
        }}
      />
    </div>
  )
}
