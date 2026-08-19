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

  const showCsv = event.schedule_mode === "fixed_range" || event.schedule_mode === "selective"

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="font-heading text-xl font-medium">Edit {event.name}</h1>
        {showCsv && (
          <div className="flex items-center gap-3 text-sm">
            <Link
              href={`/dashboard/events/${id}/days-import`}
              className="underline underline-offset-4"
            >
              Import CSV
            </Link>
            <a
              href={`/dashboard/events/${id}/days/export`}
              className="underline underline-offset-4"
            >
              Export days CSV
            </a>
          </div>
        )}
      </div>
      <EditEventForm eventId={id} event={event} />
    </div>
  )
}
