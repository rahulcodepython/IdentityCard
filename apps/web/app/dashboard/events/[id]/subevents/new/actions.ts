"use server"

import { redirect } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { createSubEvent } from "@/lib/api/subevents"
import { createSubEventSchema } from "@/lib/validation/subevents"

export type SubEventActionResult = { error: string } | undefined

export async function createSubEventAction(
  eventId: string,
  input: Record<string, unknown>
): Promise<SubEventActionResult> {
  const parsed = createSubEventSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  try {
    await createSubEvent(eventId, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${eventId}`)
}
