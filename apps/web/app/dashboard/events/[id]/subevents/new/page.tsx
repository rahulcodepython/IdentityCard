import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"

import { SubEventForm } from "../sub-event-form"
import { createSubEventAction } from "./actions"

export default async function NewSubEventPage({
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
      <p className="text-sm text-muted-foreground">
        Sub-events can only be added while the event is a draft.
      </p>
    )
  }

  const action = createSubEventAction.bind(null, id)

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">New sub-event</h1>
      <SubEventForm
        eventDays={event.days}
        action={action}
        submitLabel="Create sub-event"
      />
    </div>
  )
}
