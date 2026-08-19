"use server"

import { redirect } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { createEvent } from "@/lib/api/events"
import {
  type CreateEventInput,
  createEventSchema,
} from "@/lib/validation/events"

export type EventActionResult =
  | { error: string; code?: string; fields?: Record<string, string> }
  | undefined

export async function createEventAction(
  input: CreateEventInput
): Promise<EventActionResult> {
  const parsed = createEventSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  let event
  try {
    event = await createEvent(parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message, code: err.code, fields: err.fields }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${event.id}`)
}
