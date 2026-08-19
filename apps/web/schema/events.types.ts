import { z } from "zod"

// Mirrors apps/server/internal/modules/events/dto.go. Dates/times are
// plain "YYYY-MM-DD" / "HH:MM" strings end to end — they line up directly
// with native <input type="date"> / <input type="time"> values and with
// react-day-picker's own date formatting.

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

export const scheduleModeSchema = z.enum(["flash", "fixed_range", "selective", "recurring"])
export type ScheduleMode = z.infer<typeof scheduleModeSchema>

export const eventStatusSchema = z.enum(["draft", "published"])

export const recurrenceWeekdayInputSchema = z.object({
  weekday: z.number().min(0).max(6), // 0=Monday..6=Sunday
  entry_time: z.string().min(1, "Entry time is required"),
  exit_time: z.string().min(1, "Exit time is required"),
})
export type RecurrenceWeekdayInput = z.infer<typeof recurrenceWeekdayInputSchema>

export const recurrenceInputSchema = z.object({
  starts_on: z.string().min(1, "Start date is required"),
  ends_on: z.string().nullish(),
  weekdays: z.array(recurrenceWeekdayInputSchema).min(1, "Pick at least one weekday"),
})
export type RecurrenceInput = z.infer<typeof recurrenceInputSchema>

export const recurrenceWeekdayResponseSchema = z.object({
  weekday: z.number(),
  entry_time: z.string(),
  exit_time: z.string(),
})

export const recurrenceResponseSchema = z.object({
  starts_on: z.string(),
  ends_on: z.string().nullable(),
  weekdays: z.array(recurrenceWeekdayResponseSchema),
})

// createEventSchema is discriminated by schedule_mode — which of the
// other fields matter depends on it (see CreateScheduleInput below for
// the per-mode narrowed shapes used while authoring; this wider schema is
// what's actually posted to the API, same discriminated union the Go
// CreateEventRequest represents with plain optional fields).
export const createEventSchema = z.object({
  name: z.string().min(2, "Enter an event name").max(200),
  schedule_mode: scheduleModeSchema,
  venue: z.string().max(300).default(""),
  days: z.array(eventDayInputSchema).default([]),
  range_start: z.string().default(""),
  range_end: z.string().default(""),
  range_entry_time: z.string().default(""),
  range_exit_time: z.string().default(""),
  excluded_dates: z.array(z.string()).default([]),
  recurrence: recurrenceInputSchema.nullish(),
})
export type CreateEventInput = z.infer<typeof createEventSchema>

export const updateEventSchema = createEventSchema.omit({ schedule_mode: true })
export type UpdateEventInput = z.infer<typeof updateEventSchema>

export const eventSummarySchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  schedule_mode: scheduleModeSchema,
  status: eventStatusSchema,
  start_date: z.string(),
  end_date: z.string().nullable(), // null for an open-ended recurring event
  venue: z.string().nullable(),
  published_at: z.string().nullable(),
})
export type EventSummary = z.infer<typeof eventSummarySchema>

// events.EventResponse embeds EventSummary anonymously in Go, which
// flattens its fields into the same JSON object — so the schema extends
// rather than nests.
export const eventResponseSchema = eventSummarySchema.extend({
  days: z.array(eventDayResponseSchema),
  recurrence: recurrenceResponseSchema.nullish(),
  excluded_dates: z.array(z.string()).nullish(),
})
export type EventDetail = z.infer<typeof eventResponseSchema>

export const eventsListResponseSchema = z.array(eventSummarySchema)

export const addExcludedDateSchema = z.object({
  date: z.string().min(1, "Date is required"),
})
export type AddExcludedDateInput = z.infer<typeof addExcludedDateSchema>

export const dayImportRowErrorSchema = z.object({
  row: z.number(),
  message: z.string(),
})
export const dayImportSummarySchema = z.object({
  imported: z.number(),
  skipped: z.number(),
  errors: z.array(dayImportRowErrorSchema),
})
export type DayImportSummary = z.infer<typeof dayImportSummarySchema>
