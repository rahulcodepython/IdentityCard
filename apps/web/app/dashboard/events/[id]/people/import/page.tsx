"use client"

import { notFound, useParams } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { listSubEvents } from "@/lib/client-api/subevents"
import { queryKeys } from "@/react-query/query-keys"

import { ImportForm } from "./import-form"

export default function ImportPeoplePage() {
  const { id } = useParams<{ id: string }>()

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

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">
        Import people — {event.name}
      </h1>
      <ImportForm eventId={id} subEvents={subEvents} />
    </div>
  )
}