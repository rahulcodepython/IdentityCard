import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  type CreateEventInput,
  createEventSchema,
  eventResponseSchema,
  eventsListResponseSchema,
  type UpdateEventInput,
  updateEventSchema,
} from "@/lib/validation/events"

export async function listEvents() {
  const data = await apiFetch("/events")
  return eventsListResponseSchema.parse(data)
}

export async function getEvent(id: string) {
  const data = await apiFetch(`/events/${id}`)
  return eventResponseSchema.parse(data)
}

export async function createEvent(input: CreateEventInput) {
  const body = createEventSchema.parse(input)
  const data = await apiFetch("/events", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return eventResponseSchema.parse(data)
}

export async function updateEvent(id: string, input: UpdateEventInput) {
  const body = updateEventSchema.parse(input)
  const data = await apiFetch(`/events/${id}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  })
  return eventResponseSchema.parse(data)
}

export async function publishEvent(id: string) {
  const data = await apiFetch(`/events/${id}/publish`, { method: "POST" })
  return eventResponseSchema.parse(data)
}

export async function deleteEvent(id: string) {
  await apiFetch(`/events/${id}`, { method: "DELETE" })
}
