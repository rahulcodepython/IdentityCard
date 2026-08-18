import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"
import { listForms } from "@/lib/api/forms"
import { listSubEvents } from "@/lib/api/subevents"

import { CreateFormForm } from "./create-form-form"
import { FormRow } from "./form-row"

export default async function FormsPage({
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
  const [forms, subEvents] = await Promise.all([
    listForms(id),
    listSubEvents(id),
  ])
  const subEventNames = new Map(subEvents.map((se) => [se.id, se.name]))

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">
        Public forms — {event.name}
      </h1>
      <p className="text-sm text-muted-foreground">
        Share a link below to let people register themselves, optionally capped
        at a fixed number of submissions.
      </p>

      <CreateFormForm eventId={id} subEvents={subEvents} />

      {forms.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No public sign-up links yet.
        </p>
      ) : (
        <div className="flex flex-col gap-3">
          {forms.map((form) => (
            <FormRow
              key={form.id}
              eventId={id}
              form={form}
              subEventName={
                form.sub_event_id
                  ? subEventNames.get(form.sub_event_id)
                  : undefined
              }
            />
          ))}
        </div>
      )}
    </div>
  )
}
