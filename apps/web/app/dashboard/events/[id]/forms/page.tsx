"use client"

import { notFound, useParams } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { listForms } from "@/lib/client-api/forms"
import { listSubEvents } from "@/lib/client-api/subevents"
import { queryKeys } from "@/react-query/query-keys"

import { CreateFormForm } from "./create-form-form"
import { FormRow } from "./form-row"

export default function FormsPage() {
  const { id } = useParams<{ id: string }>()

  const { data: event, error } = useQuery({
    queryKey: queryKeys.event(id),
    queryFn: () => getEvent(id),
    retry: false,
  })

  const { data: forms = [] } = useQuery({
    queryKey: queryKeys.forms(id),
    queryFn: () => listForms(id),
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