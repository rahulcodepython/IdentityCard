import { z } from "zod"

// Mirrors apps/server/internal/modules/analytics/dto.go.

export const statusCountsSchema = z.object({
  early: z.number(),
  on_time: z.number(),
  late: z.number(),
})

export const dailyBreakdownSchema = z.object({
  date: z.string(),
  expected: z.number(),
  present: z.number(),
  absent: z.number(),
  entry_statuses: statusCountsSchema,
  exit_statuses: statusCountsSchema,
})
export type DailyBreakdown = z.infer<typeof dailyBreakdownSchema>

export const subEventSummarySchema = z.object({
  sub_event_id: z.string().uuid(),
  sub_event_name: z.string(),
  total_people: z.number(),
  attended: z.number(),
  absent: z.number(),
})
export type SubEventSummary = z.infer<typeof subEventSummarySchema>

export const summaryResponseSchema = z.object({
  total_people: z.number(),
  attended: z.number(),
  absent: z.number(),
  sub_events: z.array(subEventSummarySchema),
})
export type AnalyticsSummary = z.infer<typeof summaryResponseSchema>

export const dailyResponseSchema = z.object({
  days: z.array(dailyBreakdownSchema),
})
