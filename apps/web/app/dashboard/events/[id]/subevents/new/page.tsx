"use client"

import { notFound, useParams, useRouter } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { createSubEvent } from "@/lib/client-api/subevents"
import { queryKeys } from "@/react-query/query-keys"

import { SubEventForm } from "../sub-event-form"

export default function NewSubEventPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()

  const { data: event, error } = useQuery({
    queryKey: queryKeys.event(id),
    queryFn: () => getEvent(id),
    retry: false,
  })

  if (error instanceof ApiError && error.status === 404) notFound()

  if (!event) {
    return (
      <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
        Loading…
      </div>
    )
  }

  if (event.status !== "draft") {
    return (
      <p className="text-sm text-muted-foreground">
        Sub-events can only be added while the event is a draft.
      </p>
    )
  }

  async function action(
    input: Record<string, unknown>
  ): Promise<{ error: string } | undefined> {
    try {
      await createSubEvent(id, input as Parameters<typeof createSubEvent>[1])
    } catch (err: any) {
      return { error: err.message ?? "Something went wrong." }
    }
    router.push(`/dashboard/events/${id}`)
  }

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