import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"

import { ImportForm } from "./import-form"

export default async function DaysImportPage({
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
