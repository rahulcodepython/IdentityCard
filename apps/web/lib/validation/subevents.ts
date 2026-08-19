import { z } from "zod"

import {
  eventDayInputSchema,
  eventDayResponseSchema,
} from "@/lib/validation/events"

// Mirrors apps/server/internal/modules/subevents/dto.go. A sub-event gets
// two of an event's four schedule modes — fixed_range and selective —
// never recurring or flash. An empty `days` list (selective mode) means
// the sub-event runs on every day of the parent event.

export const subEventScheduleModeSchema = z.enum(["fixed_range", "selective"])
export type SubEventScheduleMode = z.infer<typeof subEventScheduleModeSchema>

export const createSubEventSchema = z.object({
  name: z.string().min(2, "Enter a sub-event name").max(200),
  schedule_mode: subEventScheduleModeSchema,
  days: z.array(eventDayInputSchema).default([]),
  range_start: z.string().default(""),
  range_end: z.string().default(""),
  range_entry_time: z.string().default(""),
  range_exit_time: z.string().default(""),
})
export type CreateSubEventInput = z.infer<typeof createSubEventSchema>

export const updateSubEventSchema = createSubEventSchema.omit({ schedule_mode: true })
export type UpdateSubEventInput = z.infer<typeof updateSubEventSchema>

export const subEventResponseSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  schedule_mode: subEventScheduleModeSchema,
  days: z.array(eventDayResponseSchema),
})
export type SubEvent = z.infer<typeof subEventResponseSchema>

export const subEventsListResponseSchema = z.array(subEventResponseSchema)
