"use client"

import { useQuery } from "@tanstack/react-query"

import { listEvents } from "@/lib/client-api/events"
import { listOrgSubscriptions } from "@/lib/client-api/plans"
import { queryKeys } from "@/react-query/query-keys"

import { CreateEventDialog } from "./create-event-dialog"
import { EventsTable } from "./events-table"

export default function EventsPage() {
  const { data: events = [] } = useQuery({
    queryKey: queryKeys.events(),
    queryFn: listEvents,
  })

  const { data: subs } = useQuery({
    queryKey: queryKeys.orgSubscriptions(),
    queryFn: listOrgSubscriptions,
  })

  const allowFlash =
    subs?.subscriptions.some(
      (s) => s.kind === "flash" && s.status === "active"
    ) ?? false

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="font-heading text-xl font-medium">Events</h1>
        <CreateEventDialog allowFlash={allowFlash} />
      </div>

      <EventsTable data={events} />
    </div>
  )
}