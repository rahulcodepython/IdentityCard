import { z } from "zod"

import {
  eventDayInputSchema,
  eventDayResponseSchema,
} from "@/lib/validation/events"

// Mirrors apps/server/internal/modules/subevents/dto.go. An empty `days`
// list means the sub-event runs on every day of the parent event.

export const createSubEventSchema = z.object({
  name: z.string().min(2, "Enter a sub-event name").max(200),
  days: z.array(eventDayInputSchema).default([]),
})
export type CreateSubEventInput = z.infer<typeof createSubEventSchema>

export const updateSubEventSchema = createSubEventSchema
export type UpdateSubEventInput = z.infer<typeof updateSubEventSchema>

export const subEventResponseSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  days: z.array(eventDayResponseSchema),
})
export type SubEvent = z.infer<typeof subEventResponseSchema>

export const subEventsListResponseSchema = z.array(subEventResponseSchema)
