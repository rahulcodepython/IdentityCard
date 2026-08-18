"use server"

import { redirect } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { deleteEvent, publishEvent } from "@/lib/api/events"

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
