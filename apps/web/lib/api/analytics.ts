import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  dailyResponseSchema,
  summaryResponseSchema,
} from "@/lib/validation/analytics"

export async function getAnalyticsSummary(eventId: string) {
  const data = await apiFetch(`/events/${eventId}/analytics/summary`)
  return summaryResponseSchema.parse(data)
}

export async function getAnalyticsDaily(eventId: string, subEventId?: string) {
  const qs = subEventId ? `?sub_event_id=${subEventId}` : ""
  const data = await apiFetch(`/events/${eventId}/analytics/daily${qs}`)
  return dailyResponseSchema.parse(data)
}
