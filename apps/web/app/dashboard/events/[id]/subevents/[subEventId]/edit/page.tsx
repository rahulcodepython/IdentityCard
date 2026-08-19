import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"
import { getSubEvent } from "@/lib/api/subevents"

import { SubEventForm } from "../../sub-event-form"
import { deleteSubEventAction, updateSubEventAction } from "./actions"
import { DeleteSubEventButton } from "./delete-sub-event-button"

export default async function EditSubEventPage({
  params,
}: {
  params: Promise<{ id: string; subEventId: string }>
}) {
  const { id, subEventId } = await params

  let event
  try {
    event = await getEvent(id)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }

  let subEvent
  try {
    subEvent = await getSubEvent(id, subEventId)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }

  if (event.status !== "draft") {
    return (
      <p className="text-sm text-muted-foreground">
        Sub-events can only be edited while the event is a draft.
      </p>
    )
  }

  const updateAction = updateSubEventAction.bind(null, id, subEventId)
  const deleteAction = deleteSubEventAction.bind(null, id, subEventId)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="font-heading text-xl font-medium">
          Edit {subEvent.name}
        </h1>
        <DeleteSubEventButton action={deleteAction} />
      </div>
      <SubEventForm
        eventDays={event.days}
        defaultValues={{
          name: subEvent.name,
          scheduleMode: subEvent.schedule_mode,
          days: subEvent.days,
        }}
        action={updateAction}
        submitLabel="Save changes"
      />
    </div>
  )
}
