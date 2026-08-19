"use client"

import { notFound, useParams, useRouter } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { createPerson } from "@/lib/client-api/people"
import { listSubEvents } from "@/lib/client-api/subevents"
import { queryKeys } from "@/react-query/query-keys"

import { PersonForm } from "../person-form"

export default function NewPersonPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()

  const { data: event, error } = useQuery({
    queryKey: queryKeys.event(id),
    queryFn: () => getEvent(id),
    retry: false,
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

  async function action(values: Parameters<typeof createPerson>[1]) {
    try {
      await createPerson(id, values)
    } catch (err: any) {
      return { error: err.message ?? "Something went wrong." }
    }
    router.push(`/dashboard/events/${id}/people`)
  }

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