import { z } from "zod"

// Mirrors apps/server/internal/modules/events/dto.go. Dates/times are
// plain "YYYY-MM-DD" / "HH:MM" strings end to end — they line up directly
// with native <input type="date"> / <input type="time"> values.

export const eventDayInputSchema = z.object({
  date: z.string().min(1, "Date is required"),
  entry_time: z.string().min(1, "Entry time is required"),
  exit_time: z.string().min(1, "Exit time is required"),
})
export type EventDayInput = z.infer<typeof eventDayInputSchema>

export const eventDayResponseSchema = z.object({
  date: z.string(),
  entry_time: z.string(),
  exit_time: z.string(),
})
export type EventDay = z.infer<typeof eventDayResponseSchema>

export const eventKindSchema = z.enum(["established", "flash"])
export type EventKind = z.infer<typeof eventKindSchema>

export const eventStatusSchema = z.enum(["draft", "published"])

export const createEventSchema = z.object({
  name: z.string().min(2, "Enter an event name").max(200),
  kind: eventKindSchema,
  venue: z.string().max(300).default(""),
  days: z.array(eventDayInputSchema).min(1, "Add at least one day"),
})
export type CreateEventInput = z.infer<typeof createEventSchema>

export const updateEventSchema = z.object({
  name: z.string().min(2, "Enter an event name").max(200),
  venue: z.string().max(300).default(""),
  days: z.array(eventDayInputSchema).min(1, "Add at least one day"),
})
export type UpdateEventInput = z.infer<typeof updateEventSchema>

export const eventSummarySchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  kind: eventKindSchema,
  status: eventStatusSchema,
  start_date: z.string(),
  end_date: z.string(),
  venue: z.string().nullable(),
  published_at: z.string().nullable(),
})
export type EventSummary = z.infer<typeof eventSummarySchema>

// events.EventResponse embeds EventSummary anonymously in Go, which
// flattens its fields into the same JSON object — so the schema extends
// rather than nests.
export const eventResponseSchema = eventSummarySchema.extend({
  days: z.array(eventDayResponseSchema),
})
export type EventDetail = z.infer<typeof eventResponseSchema>

export const eventsListResponseSchema = z.array(eventSummarySchema)
