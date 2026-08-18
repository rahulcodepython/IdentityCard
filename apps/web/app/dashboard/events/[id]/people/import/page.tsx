import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"
import { listSubEvents } from "@/lib/api/subevents"

import { ImportForm } from "./import-form"

export default async function ImportPeoplePage({
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

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">
        Import people — {event.name}
      </h1>
      <ImportForm eventId={id} subEvents={subEvents} />
    </div>
  )
}
