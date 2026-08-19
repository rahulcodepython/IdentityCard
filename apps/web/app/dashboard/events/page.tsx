import { listEvents } from "@/lib/api/events"
import { listOrgSubscriptions } from "@/lib/api/plans"

import { CreateEventDialog } from "./create-event-dialog"
import { EventsTable } from "./events-table"

export default async function EventsPage() {
  const [events, subs] = await Promise.all([listEvents(), listOrgSubscriptions()])
  const allowFlash = subs.subscriptions.some((s) => s.kind === "flash" && s.status === "active")

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
