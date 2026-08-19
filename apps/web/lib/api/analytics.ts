import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  dailyResponseSchema,
  overviewResponseSchema,
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

export async function getAnalyticsOverview() {
  const data = await apiFetch("/analytics/overview")
  return overviewResponseSchema.parse(data)
}
