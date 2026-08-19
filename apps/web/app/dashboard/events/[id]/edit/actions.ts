"use server"

import { redirect } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { updateEvent } from "@/lib/api/events"
import {
  type UpdateEventInput,
  updateEventSchema,
} from "@/lib/validation/events"

export type EventActionResult =
  | { error: string; fields?: Record<string, string> }
  | undefined

export async function updateEventAction(
  eventId: string,
  input: UpdateEventInput
): Promise<EventActionResult> {
  const parsed = updateEventSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  try {
    await updateEvent(eventId, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message, fields: err.fields }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${eventId}`)
}
