"use client"

import { notFound, useParams, useRouter } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { deleteSubEvent, getSubEvent, updateSubEvent } from "@/lib/client-api/subevents"
import { queryKeys } from "@/react-query/query-keys"

import { SubEventForm } from "../../sub-event-form"
import { DeleteSubEventButton } from "./delete-sub-event-button"

export default function EditSubEventPage() {
  const { id, subEventId } = useParams<{ id: string; subEventId: string }>()
  const router = useRouter()

  const { data: event, error: eventError } = useQuery({
    queryKey: queryKeys.event(id),
    queryFn: () => getEvent(id),
    retry: false,
  })

  const { data: subEvent, error: subEventError } = useQuery({
    queryKey: queryKeys.subEvent(id, subEventId),
    queryFn: () => getSubEvent(id, subEventId),
    retry: false,
  })

  if (
    (eventError instanceof ApiError && eventError.status === 404) ||
    (subEventError instanceof ApiError && subEventError.status === 404)
  ) {
    notFound()
  }

  if (!event || !subEvent) {
    return (
      <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
        Loading…
      </div>
    )
  }

  if (event.status !== "draft") {
    return (
      <p className="text-sm text-muted-foreground">
        Sub-events can only be edited while the event is a draft.
      </p>
    )
  }

  async function updateAction(
    input: Record<string, unknown>
  ): Promise<{ error: string } | undefined> {
    try {
      await updateSubEvent(id, subEventId, input as Parameters<typeof updateSubEvent>[2])
    } catch (err: any) {
      return { error: err.message ?? "Something went wrong." }
    }
    router.push(`/dashboard/events/${id}`)
  }

  async function deleteAction(): Promise<{ error: string } | undefined> {
    try {
      await deleteSubEvent(id, subEventId)
    } catch (err: any) {
      return { error: err.message ?? "Something went wrong." }
    }
    router.push(`/dashboard/events/${id}`)
  }

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