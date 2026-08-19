import { apiRequest } from "@/react-query/client"
import {
  dailyResponseSchema,
  overviewResponseSchema,
  summaryResponseSchema,
} from "@/schema/analytics.types"

export async function getAnalyticsSummary(eventId: string) {
  return apiRequest(
    { url: `/events/${eventId}/analytics/summary`, method: "GET" },
    summaryResponseSchema
  )
}

export async function getAnalyticsDaily(eventId: string, subEventId?: string) {
  const qs = subEventId ? `?sub_event_id=${subEventId}` : ""
  return apiRequest(
    { url: `/events/${eventId}/analytics/daily${qs}`, method: "GET" },
    dailyResponseSchema
  )
}

export async function getAnalyticsOverview() {
  return apiRequest(
    { url: "/analytics/overview", method: "GET" },
    overviewResponseSchema
  )
}