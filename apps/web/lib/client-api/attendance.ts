import { apiRequest } from "@/react-query/client"
import { rosterResponseSchema } from "@/schema/attendance.types"

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
  return apiRequest(
    { url: `/events/${eventId}/attendance${rosterQueryString(filter)}`, method: "GET" },
    rosterResponseSchema
  )
}