"use client"

import { notFound, useParams } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"

import { ImportForm } from "./import-form"

export default function DaysImportPage() {
  const { id } = useParams<{ id: string }>()

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

  const showDaysImport =
    event.status === "draft" &&
    (event.schedule_mode === "fixed_range" || event.schedule_mode === "selective")
  const showExclusionsImport =
    event.schedule_mode === "recurring" ||
    (event.schedule_mode === "fixed_range" && event.status === "draft")

  if (!showDaysImport && !showExclusionsImport) {
    return (
      <p className="text-sm text-muted-foreground">
        CSV import isn&apos;t available for this event&apos;s schedule type or status.
      </p>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">
        Import CSV — {event.name}
      </h1>
      <ImportForm
        eventId={id}
        showDaysImport={showDaysImport}
        showExclusionsImport={showExclusionsImport}
      />
    </div>
  )
}