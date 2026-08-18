import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"
import { listSubEvents } from "@/lib/api/subevents"

import { PersonForm } from "../person-form"
import { createPersonAction } from "./actions"

export default async function NewPersonPage({
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
  const action = createPersonAction.bind(null, id)

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">
        Add person — {event.name}
      </h1>
      <PersonForm
        subEvents={subEvents}
        showIdentity
        action={action}
        submitLabel="Add person"
      />
    </div>
  )
}
