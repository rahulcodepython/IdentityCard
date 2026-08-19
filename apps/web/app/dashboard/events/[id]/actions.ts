"use server"

import { revalidatePath } from "next/cache"
import { redirect } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { addExcludedDate, deleteEvent, publishEvent } from "@/lib/api/events"

export type EventDetailActionResult = { error: string } | undefined

export async function publishEventAction(
  eventId: string
): Promise<EventDetailActionResult> {
  try {
    await publishEvent(eventId)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
}

export async function deleteEventAction(
  eventId: string
): Promise<EventDetailActionResult> {
  try {
    await deleteEvent(eventId)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
  redirect("/dashboard/events")
}

// AddExcludedDateAction is the one schedule edit allowed on a recurring
// event regardless of draft/published status (see
// events.Service.AddExcludedDate) — adds an unplanned holiday, never
// removes one.
export async function addExcludedDateAction(
  eventId: string,
  date: string
): Promise<EventDetailActionResult> {
  try {
    await addExcludedDate(eventId, { date })
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
  revalidatePath(`/dashboard/events/${eventId}`)
  return undefined
}
