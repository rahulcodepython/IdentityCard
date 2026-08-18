import "server-only"

import { apiFetch } from "@/lib/api/client"
import { rosterResponseSchema } from "@/lib/validation/attendance"

export type RosterFilter = {
  subEventId?: string
  date?: string
  attended?: boolean
  status?: string
}

export function rosterQueryString(filter: RosterFilter) {
  const params = new URLSearchParams()
  if (filter.subEventId) params.set("sub_event_id", filter.subEventId)
  if (filter.date) params.set("date", filter.date)
  if (filter.attended !== undefined)
    params.set("attended", String(filter.attended))
  if (filter.status) params.set("status", filter.status)
  const qs = params.toString()
  return qs ? `?${qs}` : ""
}

export async function listRoster(eventId: string, filter: RosterFilter = {}) {
  const data = await apiFetch(
    `/events/${eventId}/attendance${rosterQueryString(filter)}`
  )
  return rosterResponseSchema.parse(data ?? [])
}
