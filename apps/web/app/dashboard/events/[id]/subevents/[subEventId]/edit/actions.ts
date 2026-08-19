"use server"

import { redirect } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { deleteSubEvent, updateSubEvent } from "@/lib/api/subevents"
import { updateSubEventSchema } from "@/lib/validation/subevents"

export type SubEventActionResult = { error: string } | undefined

export async function updateSubEventAction(
  eventId: string,
  subEventId: string,
  input: Record<string, unknown>
): Promise<SubEventActionResult> {
  const parsed = updateSubEventSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  try {
    await updateSubEvent(eventId, subEventId, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${eventId}`)
}

export async function deleteSubEventAction(
  eventId: string,
  subEventId: string
): Promise<SubEventActionResult> {
  try {
    await deleteSubEvent(eventId, subEventId)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${eventId}`)
}
